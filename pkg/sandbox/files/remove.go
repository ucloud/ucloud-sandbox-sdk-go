package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Remove deletes a file, or a directory and everything under it.
func (f *Filesystem) Remove(ctx context.Context, path string) error {
	req := &filesystem.RemoveRequest{Path: path}

	if _, err := f.conn.Filesystem.Remove(ctx, f.conn.SandboxRequest(req, f.sbx)); err != nil {
		return errdefs.FromConnect(err)
	}
	return nil
}
