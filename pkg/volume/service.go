package volume

import (
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// Service is the entry point for volume operations. Get one from
// client.Client.Volumes, or build it directly on a transport client.
//
// It is safe for concurrent use.
type Service struct {
	api *api.ClientWithResponses
}

// NewService returns a Service backed by t.
func NewService(t *transport.Client) (*Service, error) {
	c, err := api.NewClientWithResponses(t.APIURL(),
		api.WithHTTPClient(t.HTTPClient()),
		api.WithRequestEditorFn(t.Authorize),
	)
	if err != nil {
		return nil, err
	}
	return &Service{api: c}, nil
}
