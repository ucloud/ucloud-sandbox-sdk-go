// Package content reads and writes files in a volume through the independent
// volumecontent.{domain} service.
// Paths accept relative and absolute forms within the volume: "file.txt" and
// "/file.txt" refer to the same entry, while "." and "/" refer to the volume
// root. Responses contain normalized relative paths.
package content

import (
	"context"
	"io"
	"net/http"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/content/api"
)

// Content provides file and directory operations for one volume. Get one with
// New when using the environment configuration, or NewWithTransport to reuse
// an existing SDK client configuration.
type Content struct {
	volumeID api.VolumeID
	api      *api.ClientWithResponses
}

// New creates a content client for volumeID using the default transport
// configuration, including UCLOUD_SANDBOX_API_KEY, region, and domain
// environment variables.
func New(volumeID api.VolumeID) (*Content, error) {
	t, err := transport.New(transport.Config{})
	if err != nil {
		return nil, err
	}
	return NewWithTransport(volumeID, t)
}

// NewWithTransport creates a content client for volumeID using t's
// authentication, HTTP client, and domain configuration.
func NewWithTransport(volumeID api.VolumeID, t *transport.Client) (*Content, error) {
	if t == nil {
		return nil, &errdefs.InvalidArgumentError{SandboxError: errdefs.SandboxError{
			Message: "transport client is required",
		}}
	}
	c, err := api.NewClientWithResponses(t.VolumeContentURL(),
		api.WithHTTPClient(t.HTTPClient()),
		api.WithRequestEditorFn(t.Authorize),
	)
	if err != nil {
		return nil, err
	}
	return &Content{volumeID: volumeID, api: c}, nil
}

// List returns entries in the requested directory. A nil params pointer uses
// the server defaults and lists the volume root.
func (c *Content) List(ctx context.Context, params *api.ListVolumeDirectoryParams) ([]api.EntryInfo, error) {
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

// MakeDir creates a directory using the generated CreateDirectory request.
func (c *Content) MakeDir(ctx context.Context, req api.CreateDirectory) (*api.EntryInfo, error) {
	resp, err := c.api.CreateVolumeDirectoryWithResponse(ctx, c.volumeID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ReadFile streams a file. The caller must close the returned response Body.
// Range and IfRange support partial downloads; response headers and the status
// code are available for interpreting single and multipart byte ranges.
func (c *Content) ReadFile(ctx context.Context, params *api.ReadVolumeFileParams) (*http.Response, error) {
	resp, err := c.api.ReadVolumeFile(ctx, c.volumeID, params)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		if err != nil {
			return nil, err
		}
		return nil, transport.Check(resp, body)
	}
	return resp, nil
}

// WriteFile uploads bytes to the path specified in params. The body is sent as
// application/octet-stream.
func (c *Content) WriteFile(ctx context.Context, params *api.WriteVolumeFileParams, body io.Reader) (*api.EntryInfo, error) {
	resp, err := c.api.WriteVolumeFileWithBodyWithResponse(ctx, c.volumeID, params,
		"application/octet-stream", body)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Stat returns metadata for the path specified in params.
func (c *Content) Stat(ctx context.Context, params *api.StatVolumePathParams) (*api.EntryInfo, error) {
	resp, err := c.api.StatVolumePathWithResponse(ctx, c.volumeID, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateMetadata changes path metadata using the generated UpdatePath request.
func (c *Content) UpdateMetadata(ctx context.Context, req api.UpdatePath) (*api.EntryInfo, error) {
	resp, err := c.api.UpdateVolumePathWithResponse(ctx, c.volumeID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Remove deletes the path specified in params recursively.
func (c *Content) Remove(ctx context.Context, params *api.DeleteVolumePathParams) error {
	resp, err := c.api.DeleteVolumePathWithResponse(ctx, c.volumeID, params)
	if err != nil {
		return err
	}
	return transport.Check(resp.HTTPResponse, resp.Body)
}

// Rename moves a path using the generated RenamePath request.
func (c *Content) Rename(ctx context.Context, req api.RenamePath) (*api.EntryInfo, error) {
	resp, err := c.api.RenameVolumePathWithResponse(ctx, c.volumeID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
