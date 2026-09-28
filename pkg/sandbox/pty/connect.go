package pty

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
)

// Connect attaches to a pseudo-terminal already open in the sandbox, so it can
// be driven from a different client than the one that created it, or picked
// up again after the stream to it dropped.
//
// Output produced while no stream was attached is not replayed. A terminal
// whose process has exited is reported as a NotFound error.
func (p *Pty) Connect(ctx context.Context, pid int, opts commands.Options) (*Handle, error) {
	req := &process.ConnectRequest{Process: commands.SelectorForPID(pid)}

	stream, err := p.conn.Process.Connect(ctx, p.conn.SandboxRequest(req, p.user))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}

	handle := newPtyHandle(p, pid)
	if err := handle.start(commands.ConnectStream{ServerStreamForClient: stream}); err != nil {
		return nil, err
	}

	return handle, nil
}
