package files

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Watcher is a polled directory watch, the non-streaming alternative to Watch.
//
// Watch holds a long-lived stream, which is the better fit when one is
// available. A Watcher instead parks state inside envd and hands out events on
// demand, which survives a client that cannot keep a connection open: a
// serverless function, or anything behind a proxy that cuts idle streams.
type Watcher struct {
	// ID identifies the watcher to envd.
	ID string

	fs *Filesystem
}

// CreateWatcher starts a polled watch on a directory.
//
// The watcher lives inside the sandbox until RemoveWatcher, so a caller that
// abandons one leaves it running.
func (f *Filesystem) CreateWatcher(ctx context.Context, path string, opts WatchOptions) (*Watcher, error) {
	if opts.Recursive && !f.conn.Version.Supports(envd.VersionRecursiveWatch) {
		return nil, &errdefs.SandboxError{
			Message: "recursive watching needs envd 0.1.4 or newer; rebuild the template",
		}
	}

	req := &filesystem.CreateWatcherRequest{
		Path:               path,
		Recursive:          opts.Recursive,
		IncludeEntry:       opts.IncludeEntry,
		AllowNetworkMounts: opts.AllowNetworkMounts,
	}

	resp, err := f.conn.Filesystem.CreateWatcher(ctx, f.conn.SandboxRequest(req, f.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}
	return &Watcher{ID: resp.Msg.GetWatcherId(), fs: f}, nil
}

// Events returns the changes seen since the last call. The result is never nil,
// and is empty when nothing has happened.
func (w *Watcher) Events(ctx context.Context) ([]*filesystem.FilesystemEvent, error) {
	req := &filesystem.GetWatcherEventsRequest{WatcherId: w.ID}

	resp, err := w.fs.conn.Filesystem.GetWatcherEvents(ctx, w.fs.conn.SandboxRequest(req, w.fs.sbx))
	if err != nil {
		return nil, errdefs.FromConnect(err)
	}
	return resp.Msg.GetEvents(), nil
}

// Remove stops the watcher and frees it inside the sandbox.
func (w *Watcher) Remove(ctx context.Context) error {
	req := &filesystem.RemoveWatcherRequest{WatcherId: w.ID}

	if _, err := w.fs.conn.Filesystem.RemoveWatcher(ctx, w.fs.conn.SandboxRequest(req, w.fs.sbx)); err != nil {
		return errdefs.FromConnect(err)
	}
	return nil
}
