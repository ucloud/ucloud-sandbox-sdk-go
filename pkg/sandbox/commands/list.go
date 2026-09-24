package commands

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// List returns the processes envd is tracking inside the sandbox. The result is
// never nil.
func (c *Commands) List(ctx context.Context) ([]ProcessInfo, error) {
	resp, err := c.conn.Process.List(ctx, c.conn.SandboxRequest(&process.ListRequest{}, c.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	listed := resp.Msg.GetProcesses()
	processes := make([]ProcessInfo, 0, len(listed))
	for _, p := range listed {
		config := p.GetConfig()
		processes = append(processes, ProcessInfo{
			PID:  int(p.GetPid()),
			Tag:  p.GetTag(),
			Cmd:  config.GetCmd(),
			Args: config.GetArgs(),
			Envs: config.GetEnvs(),
			Cwd:  config.GetCwd(),
		})
	}
	return processes, nil
}
