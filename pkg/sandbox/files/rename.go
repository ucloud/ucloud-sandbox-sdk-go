package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Rename moves an entry and returns its metadata at the new path.
func (f *Filesystem) Rename(ctx context.Context, oldPath, newPath string) (*filesystem.EntryInfo, error) {
	req := &filesystem.MoveRequest{Source: oldPath, Destination: newPath}

	resp, err := f.conn.Filesystem.Move(ctx, f.conn.SandboxRequest(req, f.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}
	return resp.Msg.GetEntry(), nil
}
