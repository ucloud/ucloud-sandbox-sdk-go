package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// List returns the entries under a directory. The result is never nil.
//
// Set FileOptions.Depth to descend further than the immediate children.
func (f *Filesystem) List(ctx context.Context, path string, depth uint32) ([]*filesystem.EntryInfo, error) {
	req := &filesystem.ListDirRequest{Path: path, Depth: depth}

	resp, err := f.conn.Filesystem.ListDir(ctx, f.conn.SandboxRequest(req, f.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}
	return resp.Msg.GetEntries(), nil
}
