package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const filesystemServiceName = "filesystem.Filesystem"

type Filesystem struct {
	sandbox *Sandbox
	rpc     *RPC
}

func newFilesystem(s *Sandbox, rpc *RPC) *Filesystem {
	return &Filesystem{sandbox: s, rpc: rpc}
}

func (f *Filesystem) Read(ctx context.Context, path string, opts ...FilesystemOption) (string, error) {
	data, err := f.ReadBytes(ctx, path, opts...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Filesystem) ReadBytes(ctx context.Context, path string, opts ...FilesystemOption) ([]byte, error) {
	rc, err := f.ReadStream(ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (f *Filesystem) ReadStream(ctx context.Context, path string, opts ...FilesystemOption) (io.ReadCloser, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	requestURL := f.sandbox.envdAPIURL + "/files?path=" + url.QueryEscape(path)
	if user != "" {
		requestURL += "&username=" + url.QueryEscape(user)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	f.sandbox.setSandboxHeaders(req)
	resp, err := f.sandbox.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPError(resp.StatusCode, string(body))
	}
	return resp.Body, nil
}

func (f *Filesystem) Write(ctx context.Context, path string, data any, opts ...FilesystemOption) (*WriteInfo, error) {
	infos, err := f.WriteFiles(ctx, []WriteEntry{{Path: path, Data: data}}, opts...)
	if err != nil {
		return nil, err
	}
	if len(infos) > 0 {
		return &infos[0], nil
	}
	return &WriteInfo{Path: path}, nil
}

func (f *Filesystem) WriteFiles(ctx context.Context, files []WriteEntry, opts ...FilesystemOption) ([]WriteInfo, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for _, entry := range files {
		part, err := writer.CreateFormFile("file", entry.Path)
		if err != nil {
			return nil, err
		}
		switch v := entry.Data.(type) {
		case string:
			_, _ = part.Write([]byte(v))
		case []byte:
			_, _ = part.Write(v)
		case io.Reader:
			_, _ = io.Copy(part, v)
		default:
			return nil, fmt.Errorf("unsupported data type: %T", entry.Data)
		}
	}
	writer.Close()
	uploadURL := f.sandbox.envdAPIURL + "/files"
	sep := "?"
	if len(files) == 1 {
		uploadURL += sep + "path=" + url.QueryEscape(files[0].Path)
		sep = "&"
	}
	if user != "" {
		uploadURL += sep + "username=" + url.QueryEscape(user)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	f.sandbox.setSandboxHeaders(req)
	resp, err := f.sandbox.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPError(resp.StatusCode, string(body))
	}
	var infos []WriteInfo
	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		infos = make([]WriteInfo, len(files))
		for i, entry := range files {
			infos[i] = WriteInfo{Path: entry.Path}
		}
	}
	return infos, nil
}

func (f *Filesystem) WriteStream(ctx context.Context, path string, reader io.Reader, opts ...FilesystemOption) (*WriteInfo, error) {
	return f.Write(ctx, path, reader, opts...)
}

type listDirRequest struct {
	Path  string `json:"path"`
	Depth int    `json:"depth,omitempty"`
}
type listDirResponse struct {
	Entries []entryInfoRaw `json:"entries"`
}
type entryInfoRaw struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Type          string `json:"type"`
	Size          string `json:"size"`
	Permissions   string `json:"permissions"`
	Mode          int    `json:"mode"`
	Owner         string `json:"owner"`
	Group         string `json:"group"`
	ModifiedTime  string `json:"modifiedTime"`
	SymlinkTarget string `json:"symlinkTarget,omitempty"`
}

func (r *entryInfoRaw) toEntryInfo() EntryInfo {
	info := EntryInfo{Name: r.Name, Path: r.Path, Type: mapFileType(r.Type), Permissions: r.Permissions, Mode: uint32(r.Mode), Owner: r.Owner, Group: r.Group}
	if r.Size != "" {
		info.Size, _ = strconv.ParseInt(r.Size, 10, 64)
	}
	if r.ModifiedTime != "" {
		info.ModifiedTime, _ = time.Parse(time.RFC3339Nano, r.ModifiedTime)
	}
	if r.SymlinkTarget != "" {
		info.SymlinkTarget = &r.SymlinkTarget
	}
	return info
}

func mapFileType(ft string) EntryType {
	if ft == "FILE_TYPE_DIRECTORY" {
		return EntryTypeDir
	}
	return EntryTypeFile
}

type statRequest struct{ Path string `json:"path"` }
type statResponse struct{ Entry entryInfoRaw `json:"entry"` }
type makeDirRequest struct{ Path string `json:"path"` }
type removeRequest struct{ Path string `json:"path"` }
type moveRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}
type moveResponse struct{ Entry entryInfoRaw `json:"entry"` }

func (f *Filesystem) List(ctx context.Context, path string, opts ...FilesystemOption) ([]EntryInfo, error) {
	cfg := f.applyOpts(opts)
	depth := cfg.depth
	if depth <= 0 {
		depth = 1
	}
	user := f.sandbox.resolveUsername(cfg.user)
	var resp listDirResponse
	if err := f.rpc.CallUnary(ctx, filesystemServiceName, "ListDir", listDirRequest{Path: path, Depth: depth}, &resp, buildAuthHeader(user)); err != nil {
		return nil, err
	}
	entries := make([]EntryInfo, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = e.toEntryInfo()
	}
	return entries, nil
}

func (f *Filesystem) MakeDir(ctx context.Context, path string, opts ...FilesystemOption) (bool, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	if err := f.rpc.CallUnary(ctx, filesystemServiceName, "MakeDir", makeDirRequest{Path: path}, nil, buildAuthHeader(user)); err != nil {
		if isRPCCode(err, RPCCodeAlreadyExists) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *Filesystem) Exists(ctx context.Context, path string, opts ...FilesystemOption) (bool, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	var resp statResponse
	err := f.rpc.CallUnary(ctx, filesystemServiceName, "Stat", statRequest{Path: path}, &resp, buildAuthHeader(user))
	if err != nil {
		if isRPCCode(err, RPCCodeNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *Filesystem) GetInfo(ctx context.Context, path string, opts ...FilesystemOption) (*EntryInfo, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	var resp statResponse
	if err := f.rpc.CallUnary(ctx, filesystemServiceName, "Stat", statRequest{Path: path}, &resp, buildAuthHeader(user)); err != nil {
		return nil, err
	}
	info := resp.Entry.toEntryInfo()
	return &info, nil
}

func (f *Filesystem) Remove(ctx context.Context, path string, opts ...FilesystemOption) error {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	return f.rpc.CallUnary(ctx, filesystemServiceName, "Remove", removeRequest{Path: path}, nil, buildAuthHeader(user))
}

func (f *Filesystem) Rename(ctx context.Context, oldPath, newPath string, opts ...FilesystemOption) (*EntryInfo, error) {
	cfg := f.applyOpts(opts)
	user := f.sandbox.resolveUsername(cfg.user)
	var resp moveResponse
	if err := f.rpc.CallUnary(ctx, filesystemServiceName, "Move", moveRequest{Source: oldPath, Destination: newPath}, &resp, buildAuthHeader(user)); err != nil {
		return nil, err
	}
	info := resp.Entry.toEntryInfo()
	return &info, nil
}

type watchDirRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}
type watchDirResponse struct {
	Start      *struct{}                `json:"start,omitempty"`
	Filesystem *watchDirFilesystemEvent `json:"filesystem,omitempty"`
}
type watchDirFilesystemEvent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (f *Filesystem) WatchDir(ctx context.Context, path string, onEvent func(FilesystemEvent), opts ...FilesystemOption) (*WatchHandle, error) {
	cfg := f.applyOpts(opts)
	if cfg.recursive && !f.sandbox.supportsRecursiveWatch() {
		return nil, &SandboxError{Message: "recursive watch requires envd >= 0.1.4"}
	}
	user := f.sandbox.resolveUsername(cfg.user)
	timeoutMs := 0
	if cfg.watchTimeout > 0 {
		timeoutMs = cfg.watchTimeout * 1000
	}
	watchCtx, cancel := context.WithCancel(ctx)
	stream, err := f.rpc.CallServerStream(watchCtx, filesystemServiceName, "WatchDir", watchDirRequest{Path: path, Recursive: cfg.recursive}, timeoutMs, buildAuthHeader(user))
	if err != nil {
		cancel()
		return nil, err
	}
	var firstEvent watchDirResponse
	if err := stream.Next(&firstEvent); err != nil {
		stream.Close()
		cancel()
		return nil, err
	}
	handle := &WatchHandle{cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(handle.done)
		defer stream.Close()
		defer cancel()
		for {
			var event watchDirResponse
			if err := stream.Next(&event); err != nil {
				if err != io.EOF {
					handle.mu.Lock()
					handle.err = err
					handle.mu.Unlock()
					if cfg.onExit != nil {
						cfg.onExit(err)
					}
				}
				return
			}
			if event.Filesystem != nil {
				onEvent(FilesystemEvent{Name: event.Filesystem.Name, Type: mapEventType(event.Filesystem.Type)})
			}
		}
	}()
	return handle, nil
}

func (f *Filesystem) applyOpts(opts []FilesystemOption) *filesystemConfig {
	cfg := &filesystemConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

type WatchHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	err    error
}

func (w *WatchHandle) Stop()                        { w.cancel() }
func (w *WatchHandle) Wait(ctx context.Context) error {
	select {
	case <-w.done:
		w.mu.Lock()
		defer w.mu.Unlock()
		return w.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func mapEventType(raw string) FilesystemEventType {
	switch raw {
	case "EVENT_TYPE_CREATE":
		return EventTypeCreate
	case "EVENT_TYPE_WRITE":
		return EventTypeWrite
	case "EVENT_TYPE_REMOVE":
		return EventTypeRemove
	case "EVENT_TYPE_RENAME":
		return EventTypeRename
	case "EVENT_TYPE_CHMOD":
		return EventTypeChmod
	default:
		return FilesystemEventType(strings.ToLower(strings.TrimPrefix(raw, "EVENT_TYPE_")))
	}
}
