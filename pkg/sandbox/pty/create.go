package pty

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
)

// Create opens a pseudo-terminal running a login shell, sized as given.
//
// The caller must drain PtyHandle.Output, or the terminal stalls once its
// buffer fills.
func (p *Pty) Create(ctx context.Context, size PtySize, opts commands.Options) (*PtyHandle, error) {
	config := &process.ProcessConfig{
		Cmd:  "/bin/bash",
		Args: []string{"-i", "-l"},
		Envs: map[string]string{"TERM": "xterm-256color"},
	}
	for key, value := range opts.EnvVars {
		config.Envs[key] = value
	}
	if opts.Cwd != "" {
		config.Cwd = &opts.Cwd
	}

	stdin := true
	req := &process.StartRequest{
		Process: config,
		Pty:     ptySize(size),
		Stdin:   &stdin,
	}

	stream, err := p.conn.Process.Start(ctx, p.conn.SandboxRequest(req, p.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	handle := newPtyHandle(p, 0)
	go handle.consume(commands.StartStream{ServerStreamForClient: stream})

	return handle, nil
}
