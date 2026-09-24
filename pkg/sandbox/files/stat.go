package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// GetInfo returns an entry's metadata.
func (f *Filesystem) GetInfo(ctx context.Context, path string) (*filesystem.EntryInfo, error) {
	req := &filesystem.StatRequest{Path: path}

	resp, err := f.conn.Filesystem.Stat(ctx, f.conn.SandboxRequest(req, f.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}
	return resp.Msg.GetEntry(), nil
}

// Exists reports whether a path exists. Failures other than "not found" are
// returned, so a network problem is not mistaken for absence.
func (f *Filesystem) Exists(ctx context.Context, path string) (bool, error) {
	if _, err := f.GetInfo(ctx, path); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
