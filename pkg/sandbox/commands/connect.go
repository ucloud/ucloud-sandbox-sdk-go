package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Connect attaches to a process already running in the sandbox, so its output
// can be followed from a different client than the one that started it.
//
// Output produced before connecting is not replayed.
func (c *Commands) Connect(ctx context.Context, pid int, opts Options) (*Handle, error) {
	req := &process.ConnectRequest{Process: SelectorForPID(pid)}

	stream, err := c.conn.Process.Connect(ctx, c.conn.SandboxRequest(req, c.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	handle := &Handle{PID: pid, cmds: c, done: make(chan struct{})}
	go handle.consume(ConnectStream{stream}, opts)

	return handle, nil
}
