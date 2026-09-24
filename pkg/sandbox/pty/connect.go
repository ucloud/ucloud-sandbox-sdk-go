package pty

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
)

// Connect attaches to a pseudo-terminal already open in the sandbox, so it can
// be driven from a different client than the one that created it.
func (p *Pty) Connect(ctx context.Context, pid int, opts commands.Options) (*PtyHandle, error) {
	req := &process.ConnectRequest{Process: commands.SelectorForPID(pid)}

	stream, err := p.conn.Process.Connect(ctx, p.conn.SandboxRequest(req, p.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	handle := newPtyHandle(p, pid)
	go handle.consume(commands.ConnectStream{ServerStreamForClient: stream})

	return handle, nil
}
