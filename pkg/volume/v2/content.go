package volume

import (
	"context"
	"io"
	"io/fs"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/contentapi"
)

// Content reads and writes the files of one volume through the volume content
// API, served from volumecontent.{domain}. Get one from Service.Content or
// client.Client.VolumeContentV2. It is safe for concurrent use.
//
// Every path is relative to the volume root; "/" names the root itself.
type Content struct {
	api      *contentapi.ClientWithResponses
	volumeID contentapi.VolumeID
}

// Content returns a handle on the files of the volume with the given UUID. It
// makes no request; a missing volume surfaces on the first call.
func (s *Service) Content(volumeID api.VolumeID) *Content {
	return &Content{api: s.content, volumeID: volumeID}
}

// VolumeID returns the UUID of the volume this handle operates on.
func (c *Content) VolumeID() api.VolumeID { return c.volumeID }

// CreateOptions controls how WriteFile and MakeDir create an entry. A nil
// *CreateOptions uses the server defaults.
type CreateOptions struct {
	// Force overwrites an existing file and creates missing parent
	// directories. For MakeDir it also accepts a directory that already exists.
	Force bool

	// UID and GID set the owner of the created entry.
	UID *int64
	GID *int64

	// Mode sets the permission bits, at most 0o777.
	Mode *fs.FileMode
}

func (o *CreateOptions) params() (force *bool, uid, gid *int64, mode *int) {
	if o == nil {
		return nil, nil, nil, nil
	}
	if o.Force {
		force = new(true)
	}
	if o.Mode != nil {
		mode = new(int(*o.Mode))
	}
	return force, o.UID, o.GID, mode
}

// ReadFile downloads the file at path. The caller must close the returned
// reader. The body is streamed, so it is subject to the HTTP client timeout
// (transport.Config.RequestTimeout) for the whole download.
//
// GET /v2/volumecontent/{volumeID}/file
func (c *Content) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := c.api.ReadVolumeFile(ctx, c.volumeID, &contentapi.ReadVolumeFileParams{Path: path})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, errdefs.FromHTTP(resp.StatusCode, string(body))
	}
	return resp.Body, nil
}

// WriteFile uploads r as the file at path and returns the created entry.
// Without opts.Force an existing file or a missing parent is an error. opts
// may be nil.
//
// A *bytes.Reader, *bytes.Buffer or *strings.Reader body can be replayed, so a
// rate-limited upload is retried; any other reader is sent once.
//
// PUT /v2/volumecontent/{volumeID}/file
func (c *Content) WriteFile(ctx context.Context, path string, r io.Reader, opts *CreateOptions) (*contentapi.Entry, error) {
	params := &contentapi.WriteVolumeFileParams{Path: path}
	params.Force, params.Uid, params.Gid, params.Mode = opts.params()
	resp, err := c.api.WriteVolumeFileWithBodyWithResponse(ctx, c.volumeID, params, "application/octet-stream", r)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// MakeDir creates the directory at path and returns it. Without opts.Force an
// existing entry or a missing parent is an error. opts may be nil.
//
// POST /v2/volumecontent/{volumeID}/dir
func (c *Content) MakeDir(ctx context.Context, path string, opts *CreateOptions) (*contentapi.Entry, error) {
	params := &contentapi.CreateVolumeDirectoryParams{Path: path}
	params.Force, params.Uid, params.Gid, params.Mode = opts.params()
	resp, err := c.api.CreateVolumeDirectoryWithResponse(ctx, c.volumeID, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// ListDir lists the directory at path. depth, from 1 to 10, is how many levels
// to descend; 0 uses the server default of 1.
//
// GET /v2/volumecontent/{volumeID}/dir
func (c *Content) ListDir(ctx context.Context, path string, depth int) ([]contentapi.Entry, error) {
	params := &contentapi.ListVolumeDirectoryParams{Path: path}
	if depth != 0 {
		params.Depth = &depth
	}
	resp, err := c.api.ListVolumeDirectoryWithResponse(ctx, c.volumeID, params)
	if err != nil {
		return nil, err
	}
	entries, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return *entries, nil
}

// Stat returns the metadata of the file, directory or symlink at path.
//
// GET /v2/volumecontent/{volumeID}/path
func (c *Content) Stat(ctx context.Context, path string) (*contentapi.Entry, error) {
	resp, err := c.api.StatVolumePathWithResponse(ctx, c.volumeID, &contentapi.StatVolumePathParams{Path: path})
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Exists reports whether path exists. A 404 is false with no error.
func (c *Content) Exists(ctx context.Context, path string) (bool, error) {
	resp, err := c.api.StatVolumePathWithResponse(ctx, c.volumeID, &contentapi.StatVolumePathParams{Path: path})
	if err != nil {
		return false, err
	}
	if transport.IsNotFound(resp.HTTPResponse) {
		return false, nil
	}
	if err := transport.Check(resp.HTTPResponse, resp.Body); err != nil {
		return false, err
	}
	return true, nil
}

// UpdateMetadata changes the owner or permission bits of path and returns the
// updated entry. At least one field of md must be set.
//
// PATCH /v2/volumecontent/{volumeID}/path
func (c *Content) UpdateMetadata(ctx context.Context, path string, md contentapi.Metadata) (*contentapi.Entry, error) {
	resp, err := c.api.UpdateVolumeMetadataWithResponse(ctx, c.volumeID,
		&contentapi.UpdateVolumeMetadataParams{Path: path}, md)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Rename moves a file or directory from params.Path to the exact params.NewPath
// within the volume and returns its metadata at the new path. params must be
// non-nil. Both parent directories must exist; the volume root cannot be moved
// or replaced. Set params.Force to replace an existing file or empty directory.
// The target existence check and rename are not atomic.
//
// POST /v2/volumecontent/{volumeID}/rename
func (c *Content) Rename(ctx context.Context, params *contentapi.RenameVolumePathParams) (*contentapi.Entry, error) {
	resp, err := c.api.RenameVolumePathWithResponse(ctx, c.volumeID, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Remove deletes the file or directory at path, recursively. The volume root
// cannot be removed.
//
// DELETE /v2/volumecontent/{volumeID}/path
func (c *Content) Remove(ctx context.Context, path string) error {
	resp, err := c.api.DeleteVolumePathWithResponse(ctx, c.volumeID, &contentapi.DeleteVolumePathParams{Path: path})
	if err != nil {
		return err
	}
	return transport.Check(resp.HTTPResponse, resp.Body)
}
