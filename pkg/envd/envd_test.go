package envd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process/processconnect"
)

// clientTimeout is the http.Client.Timeout the tests run under. Every handler
// below takes longer than it to finish.
const clientTimeout = 200 * time.Millisecond

// slowProcess streams a process that outlives clientTimeout, and answers
// SendInput only after it.
type slowProcess struct {
	processconnect.UnimplementedProcessHandler
}

func (slowProcess) Start(_ context.Context, _ *connect.Request[process.StartRequest], stream *connect.ServerStream[process.StartResponse]) error {
	events := []*process.ProcessEvent{
		{Event: &process.ProcessEvent_Start{Start: &process.ProcessEvent_StartEvent{Pid: 42}}},
		{Event: &process.ProcessEvent_End{End: &process.ProcessEvent_EndEvent{Exited: true}}},
	}
	for i, event := range events {
		if i > 0 {
			time.Sleep(3 * clientTimeout)
		}
		if err := stream.Send(&process.StartResponse{Event: event}); err != nil {
			return err
		}
	}
	return nil
}

func (slowProcess) SendInput(ctx context.Context, _ *connect.Request[process.SendInputRequest]) (*connect.Response[process.SendInputResponse], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(3 * clientTimeout):
	}
	return connect.NewResponse(&process.SendInputResponse{}), nil
}

func newSlowConn(t *testing.T) *Connection {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle(processconnect.NewProcessHandler(slowProcess{}))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = clientTimeout

	conn, err := newConn(httpClient, server.URL, "sbx", "", "", "key", "", Version{})
	require.NoError(t, err)
	return conn
}

func TestStreamOutlivesClientTimeout(t *testing.T) {
	conn := newSlowConn(t)

	stream, err := conn.Process.Start(context.Background(), connect.NewRequest(&process.StartRequest{}))
	require.NoError(t, err)
	defer stream.Close()

	var events []*process.ProcessEvent
	for stream.Receive() {
		events = append(events, stream.Msg().GetEvent())
	}
	require.NoError(t, stream.Err())
	require.Len(t, events, 2)
	assert.NotNil(t, events[1].GetEnd())
}

func TestUnaryKeepsClientTimeout(t *testing.T) {
	conn := newSlowConn(t)

	_, err := conn.Process.SendInput(context.Background(), connect.NewRequest(&process.SendInputRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeDeadlineExceeded, connect.CodeOf(err))
}

func TestUnaryHonoursCallerDeadline(t *testing.T) {
	conn := newSlowConn(t)

	// A caller allowing longer than the client timeout is not cut short.
	ctx, cancel := context.WithTimeout(context.Background(), 10*clientTimeout)
	defer cancel()

	_, err := conn.Process.SendInput(ctx, connect.NewRequest(&process.SendInputRequest{}))
	assert.NoError(t, err)
}
