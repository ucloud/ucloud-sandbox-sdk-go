package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// SendStdin writes to a process's stdin. The process must have been started
// with CommandOptions.Stdin set, or there is nothing listening.
func (c *Commands) SendStdin(ctx context.Context, pid int, data string) error {
	req := &process.SendInputRequest{
		Process: SelectorForPID(pid),
		Input: &process.ProcessInput{
			Input: &process.ProcessInput_Stdin{Stdin: []byte(data)},
		},
	}

	if _, err := c.conn.Process.SendInput(ctx, c.conn.SandboxRequest(req, c.sbx)); err != nil {
		return errdefs.FromConnect(err)
	}
	return nil
}
