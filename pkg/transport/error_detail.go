package transport

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// maxErrorBody bounds how much of a failed response is read looking for an
// explanation. An error body is small; anything larger is not one.
const maxErrorBody = 64 << 10

// errorDetailTransport folds the explanation out of a failed response's body
// into that response's status line.
//
// It exists for envd, which is reached over connect-rpc. When the gateway in
// front of envd rejects a request it answers in its own JSON shape rather than
// connect's, and connect-go has no use for a body it cannot decode: a failed
// stream is reported as "HTTP status 401 Unauthorized" and the body — the part
// that says *why* — is dropped. connect does keep the status line, so putting
// the message there is what carries it through to the caller.
//
// Successful responses are passed through untouched, streams included. Only a
// non-2xx response is read, and its body is restored for whatever reads it
// next, since the control plane's own error mapping still wants it.
type errorDetailTransport struct{ base http.RoundTripper }

func (t *errorDetailTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	resp.Body = restoreBody(body, resp.Body)
	if readErr != nil {
		return resp, nil
	}

	if message := errdefs.BodyMessage(string(body)); message != "" {
		resp.Status = resp.Status + ": " + message
	}
	return resp, nil
}

// restoreBody hands back a body that reads as if nothing had consumed it:
// the bytes already read, then whatever is left beyond the read limit.
func restoreBody(read []byte, rest io.ReadCloser) io.ReadCloser {
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: io.MultiReader(bytes.NewReader(read), rest),
		Closer: rest,
	}
}
