package commands

import (
	"strings"

	"connectrpc.com/connect"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// ProcessStream is the half of a connect server stream that Start and Connect
// share: both deliver process.ProcessEvent, just wrapped in different response
// types.
type ProcessStream interface {
	Receive() bool
	Err() error
	Close() error
	Event() *process.ProcessEvent
}

type StartStream struct {
	*connect.ServerStreamForClient[process.StartResponse]
}

func (s StartStream) Event() *process.ProcessEvent { return s.Msg().GetEvent() }

type ConnectStream struct {
	*connect.ServerStreamForClient[process.ConnectResponse]
}

func (s ConnectStream) Event() *process.ProcessEvent { return s.Msg().GetEvent() }

// appendOutput records a chunk of output and hands it to the callback.
//
// The callback runs while the lock is not held, so a slow or re-entrant
// callback cannot block a reader calling Stdout.
func (h *Handle) appendOutput(builder *strings.Builder, chunk []byte, callback func(string)) {
	text := string(chunk)

	h.mu.Lock()
	builder.WriteString(text)
	h.mu.Unlock()

	if callback != nil {
		callback(text)
	}
}

// consume drains a process stream into the handle, then closes it.
//
// It runs in its own goroutine for Start and Connect, and inline for Run. The
// handle's done channel is closed exactly once, when the stream ends.
func (h *Handle) consume(stream ProcessStream, opts Options) {
	defer close(h.done)
	defer stream.Close()

	for stream.Receive() {
		switch event := stream.Event().GetEvent().(type) {
		case *process.ProcessEvent_Start:
			h.PID = int(event.Start.GetPid())

		case *process.ProcessEvent_Data:
			switch data := event.Data.GetOutput().(type) {
			case *process.ProcessEvent_DataEvent_Stdout:
				h.appendOutput(&h.stdout, data.Stdout, opts.OnStdout)
			case *process.ProcessEvent_DataEvent_Stderr:
				h.appendOutput(&h.stderr, data.Stderr, opts.OnStderr)
			}

		case *process.ProcessEvent_End:
			h.result = StreamEnd(event.End, h.Stdout(), h.Stderr())

		case *process.ProcessEvent_Keepalive:
			// Only there to keep the connection from going idle.
		}
	}

	if err := stream.Err(); err != nil {
		h.err = errdefs.FromConnect(err)
		return
	}
	if h.result == nil {
		// The stream ended without an end event, so the command's fate is
		// unknown. Reporting a zero exit code here would claim success.
		h.err = StreamEndEarly()
	}
}

// StreamEnd builds the result of a finished process.
func StreamEnd(end *process.ProcessEvent_EndEvent, stdout, stderr string) *Result {
	result := &Result{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: int(end.GetExitCode()),
	}
	if end.Error != nil {
		result.Error = *end.Error
	}
	return result
}

func StreamEndEarly() error {
	return &errdefs.SandboxError{
		Message: "process stream ended without reporting an exit status",
	}
}
