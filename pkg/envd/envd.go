package envd

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	envdapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem/filesystemconnect"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process/processconnect"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

const (
	Port = 49983

	// keepalivePingIntervalSec tells envd how often to send a keep-alive on a
	// stream, so an idle stream is not mistaken for a dead one.
	keepalivePingIntervalSec = 50
)

// envd's own headers. The access token authenticates to envd itself; the
// traffic token gets past the proxy in front of it; the sandbox ID and port
// tell that proxy which sandbox to route to.
const (
	headerAccessToken        = "X-Access-Token"
	headerTrafficAccessToken = "E2B-Traffic-Access-Token"
	headerSandboxID          = "E2b-Sandbox-Id"
	headerSandboxPort        = "E2b-Sandbox-Port"
	headerKeepalivePing      = "Keepalive-Ping-Interval"
)

type Connection struct {
	Process    processconnect.ProcessClient
	Filesystem filesystemconnect.FilesystemClient
	Files      *envdapi.ClientWithResponses

	Version Version

	baseURL string
	headers map[string]string

	user string
}

func Connect(t *transport.Client, sbx *api.Sandbox, user string) (*Connection, error) {
	var accessToken string
	if sbx.EnvdAccessToken != nil {
		accessToken = *sbx.EnvdAccessToken
	}
	var trafficToken string
	if sbx.TrafficAccessToken != nil {
		trafficToken = *sbx.TrafficAccessToken
	}
	version := ParseVersion(sbx.EnvdVersion)
	return newConn(
		t.HTTPClient(),
		t.SandboxURL(sbx.SandboxID, t.Domain()),
		sbx.SandboxID,
		accessToken,
		trafficToken,
		t.APIKey(),
		user,
		version,
	)
}

// newConn builds the clients for one sandbox's envd.
//
// The connect clients use the default protobuf codec. envd speaks the connect
// protocol natively, so no per-call encoding choices are needed here.
func newConn(httpClient *http.Client, baseURL, sandboxID, accessToken, trafficToken, apiKey, user string, version Version) (*Connection, error) {
	headers := map[string]string{
		headerSandboxID:     sandboxID,
		headerSandboxPort:   strconv.Itoa(Port),
		headerKeepalivePing: strconv.Itoa(keepalivePingIntervalSec),
	}
	// A secured sandbox issues its own access token. Without one, the
	// account's API key is what envd will accept.
	if accessToken != "" {
		headers[headerAccessToken] = accessToken
	} else if apiKey != "" {
		headers["X-API-Key"] = apiKey
	}
	if trafficToken != "" {
		headers[headerTrafficAccessToken] = trafficToken
	}

	conn := &Connection{
		baseURL: baseURL,
		headers: headers,
		user:    user,
		Version: version,
	}

	interceptor := connect.WithInterceptors(headerInterceptor{conn: conn})

	conn.Process = processconnect.NewProcessClient(httpClient, baseURL, interceptor)
	conn.Filesystem = filesystemconnect.NewFilesystemClient(httpClient, baseURL, interceptor)

	filesClient, err := envdapi.NewClientWithResponses(
		baseURL,
		envdapi.WithHTTPClient(httpClient),
		envdapi.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			conn.applyHeaders(req.Header)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	conn.Files = filesClient

	return conn, nil
}

// applyHeaders adds envd's headers to an outgoing request.
func (c *Connection) applyHeaders(header http.Header) {
	for key, value := range c.headers {
		header.Set(key, value)
	}
}

// headerInterceptor puts envd's headers on every outgoing call, streaming ones
// included.
//
// connect.UnaryInterceptorFunc would be the obvious way to write this, but it
// is documented as having no effect on streaming RPCs. Process.Start,
// Process.Connect and Filesystem.WatchDir are all server streams, so with that
// interceptor they reached envd carrying no access token and were rejected
// with a 401 — which surfaced only once the stream was read, not when the call
// was made.
type headerInterceptor struct{ conn *Connection }

func (i headerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		i.conn.applyHeaders(req.Header())
		return next(ctx, req)
	}
}

func (i headerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		i.conn.applyHeaders(conn.RequestHeader())
		return conn
	}
}

// WrapStreamingHandler is the server half of the interface, which this client
// never exercises.
func (i headerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func (c *Connection) SandboxRequest[T any](msg *T, sbx *api.Sandbox) *connect.Request[T] {
	return connectRequestWithUser(connect.NewRequest(msg), c.GetUser())
}

func (c *Connection) GetUser() string {
	if c.user != "" {
		return c.user
	}
	if !c.Version.Supports(VersionDefaultUser) {
		return "user"
	}
	return ""
}

// userHeader returns the basic-auth header envd reads the acting user from, or
// nil when the sandbox's default user should be used.
//
// envd takes the user in the password-less half of a basic-auth credential,
// which is why this is not a plain header value.
func userHeader(user string) map[string]string {
	if user == "" {
		return nil
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(user + ":"))
	return map[string]string{"Authorization": "Basic " + encoded}
}

// connectRequestWithUser attaches the acting user to a connect request.
func connectRequestWithUser[T any](req *connect.Request[T], user string) *connect.Request[T] {
	for key, value := range userHeader(user) {
		req.Header().Set(key, value)
	}
	return req
}
