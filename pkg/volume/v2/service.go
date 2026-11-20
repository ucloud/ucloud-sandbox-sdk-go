// Package volume manages the cloud NFS volumes exposed by /v2/volumes.
// Sandbox mounts are configured through metadata; see WithMount.
package volume

import (
	"context"
	"net/http"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// Service manages project volumes. Get one from client.Client.VolumesV2, or
// build it directly on a transport client. It is safe for concurrent use.
type Service struct {
	api *api.ClientWithResponses
}

// NewService returns a Service sharing t's HTTP client and configuration.
func NewService(t *transport.Client) (*Service, error) {
	c, err := api.NewClientWithResponses(t.APIURL(),
		api.WithHTTPClient(t.HTTPClient()),
		api.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			for k, v := range t.Headers() {
				req.Header.Set(k, v)
			}
			req.Header.Set("X-API-Key", t.APIKey())
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Service{api: c}, nil
}
