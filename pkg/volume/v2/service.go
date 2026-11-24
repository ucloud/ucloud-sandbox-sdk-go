// Package volume manages the cloud NFS volumes exposed by /v2/volumes, and the
// files inside them through the volume content API; see Service.Content.
// Sandbox mounts are configured through metadata; see WithMount.
package volume

import (
	"context"
	"net/http"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/contentapi"
)

// Service manages project volumes. Get one from client.Client.VolumesV2, or
// build it directly on a transport client. It is safe for concurrent use.
type Service struct {
	api     *api.ClientWithResponses
	content *contentapi.ClientWithResponses
}

// NewService returns a Service sharing t's HTTP client and configuration.
// Volume management goes to t.APIURL, file operations to t.VolumeContentURL.
func NewService(t *transport.Client) (*Service, error) {
	authorize := func(_ context.Context, req *http.Request) error {
		for k, v := range t.Headers() {
			req.Header.Set(k, v)
		}
		req.Header.Set("X-API-Key", t.APIKey())
		return nil
	}

	c, err := api.NewClientWithResponses(t.APIURL(),
		api.WithHTTPClient(t.HTTPClient()),
		api.WithRequestEditorFn(authorize),
	)
	if err != nil {
		return nil, err
	}
	content, err := contentapi.NewClientWithResponses(t.VolumeContentURL(),
		contentapi.WithHTTPClient(t.HTTPClient()),
		contentapi.WithRequestEditorFn(authorize),
	)
	if err != nil {
		return nil, err
	}
	return &Service{api: c, content: content}, nil
}
