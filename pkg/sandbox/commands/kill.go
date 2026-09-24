package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Kill sends SIGKILL to a process. A process that is already gone is reported
// as false rather than as an error.
func (c *Commands) Kill(ctx context.Context, pid int) (bool, error) {
	req := &process.SendSignalRequest{
		Process: SelectorForPID(pid),
		Signal:  process.Signal_SIGNAL_SIGKILL,
	}

	if _, err := c.conn.Process.SendSignal(ctx, c.conn.SandboxRequest(req, c.sbx)); err != nil {
		mapped := errdefs.FromConnect(err)
		if errdefs.IsNotFound(mapped) {
			return false, nil
		}
		return false, mapped
	}
	return true, nil
}
