package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// CloseStdin closes a process's stdin, signalling EOF to a command that reads
// until the input runs out.
//
// This has no effect on a PTY; send Ctrl-D (0x04) there instead.
func (c *Commands) CloseStdin(ctx context.Context, pid int) error {
	req := &process.CloseStdinRequest{Process: SelectorForPID(pid)}

	if _, err := c.conn.Process.CloseStdin(ctx, c.conn.SandboxRequest(req, c.sbx)); err != nil {
		return errdefs.FromConnect(err)
	}
	return nil
}
