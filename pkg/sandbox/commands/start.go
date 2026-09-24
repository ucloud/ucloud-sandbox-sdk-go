package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Start runs a command and returns without waiting for it. Use the handle to
// read output as it arrives, feed stdin, wait, or kill.
//
// The command runs through a login shell, so pipes, redirection and globbing
// work as typed.
func (c *Commands) Start(ctx context.Context, cmd string, opts Options) (*Handle, error) {
	req := &process.StartRequest{Process: c.processConfig(cmd, opts)}

	// envd defaults stdin to open for backwards compatibility. Newer agents
	// let the client choose, and leaving it closed is the better default: a
	// command reading from a stdin nobody writes to would hang.
	envdVersion := envd.ParseVersion(c.sbx.EnvdVersion)
	if envdVersion.Supports(envd.VersionStdin) {
		stdin := opts.Stdin
		req.Stdin = &stdin
	}

	stream, err := c.conn.Process.Start(ctx, c.conn.SandboxRequest(req, c.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	handle := &Handle{cmds: c, done: make(chan struct{})}
	go handle.consume(StartStream{stream}, opts)

	return handle, nil
}
