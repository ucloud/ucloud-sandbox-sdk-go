package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// Create synchronously creates a volume. params may be nil. A successful
// idempotent replay returns the existing volume, just like a fresh creation.
//
// POST /v2/volumes
func (s *Service) Create(ctx context.Context, req api.CreateVolume, params *api.CreateVolumeParams) (*api.Volume, error) {
	resp, err := s.api.CreateVolumeWithResponse(ctx, params, req)
	if err != nil {
		return nil, err
	}
	if resp.JSON201 != nil {
		return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
