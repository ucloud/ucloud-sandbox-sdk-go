package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// MakeDir creates a directory, along with any missing parents. It reports false
// when the directory already existed.
func (f *Filesystem) MakeDir(ctx context.Context, path string) (bool, error) {
	req := &filesystem.MakeDirRequest{Path: path}

	if _, err := f.conn.Filesystem.MakeDir(ctx, f.conn.SandboxRequest(req, f.sbx)); err != nil {
		mapped := errdefs.FromConnect(err)
		if errdefs.IsConflict(mapped) {
			return false, nil
		}
		return false, mapped
	}
	return true, nil
}
