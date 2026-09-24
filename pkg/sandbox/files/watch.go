package files

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/filesystem"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// WatchOptions are the optional arguments to Filesystem.Watch and
// Filesystem.CreateWatcher.
type WatchOptions struct {
	// Recursive watches the whole tree below the path. Requires envd 0.1.4 or
	// newer.
	Recursive bool

	// IncludeEntry attaches each affected entry's metadata to its event,
	// where the entry still exists.
	IncludeEntry bool

	// AllowNetworkMounts permits watching a network filesystem, where events
	// may be unreliable or absent altogether.
	AllowNetworkMounts bool

	// TimeoutSeconds stops the watch after this long. Zero watches until the
	// handle is stopped or the context ends.
	TimeoutSeconds int

	// OnExit is called once the watch ends, with the error that ended it or
	// nil for a clean stop.
	OnExit func(error)
}

// WatchHandle is a running directory watch. Stop it, or wait for it to end.
type WatchHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

// Stop ends the watch. It is safe to call more than once.
func (w *WatchHandle) Stop() { w.cancel() }

// Wait blocks until the watch ends and returns why.
//
// A watch stopped with Stop returns nil, as does one whose timeout expired: a
// caller who asked for the watch to end is not told that it did.
func (w *WatchHandle) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return w.err
	}
}

// Watch reports changes under a directory to onEvent until the handle is
// stopped, the timeout expires or the context ends.
//
// onEvent is called from the watch's own goroutine, one call at a time, so it
// must not block for long.
func (f *Filesystem) Watch(ctx context.Context, path string, onEvent func(*filesystem.FilesystemEvent), opts WatchOptions) (*WatchHandle, error) {
	if opts.Recursive && !f.conn.Version.Supports(envd.VersionRecursiveWatch) {
		return nil, &errdefs.SandboxError{
			Message: "recursive watching needs envd 0.1.4 or newer; rebuild the template",
		}
	}

	watchCtx, cancel := context.WithCancel(ctx)
	if opts.TimeoutSeconds > 0 {
		watchCtx, cancel = context.WithTimeout(ctx, time.Duration(opts.TimeoutSeconds)*time.Second)
	}

	req := &filesystem.WatchDirRequest{
		Path:               path,
		Recursive:          opts.Recursive,
		IncludeEntry:       opts.IncludeEntry,
		AllowNetworkMounts: opts.AllowNetworkMounts,
	}

	stream, err := f.conn.Filesystem.WatchDir(watchCtx, f.conn.SandboxRequest(req, f.sbx))
	if err != nil {
		cancel()
		return nil, errdefs.FromConnect(err)
	}

	handle := &WatchHandle{cancel: cancel, done: make(chan struct{})}
	go handle.consume(watchCtx, stream, onEvent, opts.OnExit)

	return handle, nil
}

// consume drains the watch stream until it ends.
func (w *WatchHandle) consume(
	ctx context.Context,
	stream *connect.ServerStreamForClient[filesystem.WatchDirResponse],
	onEvent func(*filesystem.FilesystemEvent),
	onExit func(error),
) {
	defer close(w.done)
	defer stream.Close()
	defer w.cancel()

	for stream.Receive() {
		event, ok := stream.Msg().GetEvent().(*filesystem.WatchDirResponse_Filesystem)
		if !ok {
			// The other cases are the stream's own start and keep-alive
			// markers, which say nothing about the filesystem.
			continue
		}
		if onEvent != nil {
			if event.Filesystem != nil {
				onEvent(event.Filesystem)
			}
		}
	}

	// A watch the caller ended -- Stop, or the timeout -- is not a failure,
	// so the context's error is not reported as one.
	if err := stream.Err(); err != nil && ctx.Err() == nil {
		w.err = errdefs.FromConnect(err)
	}
	if onExit != nil {
		onExit(w.err)
	}
}
