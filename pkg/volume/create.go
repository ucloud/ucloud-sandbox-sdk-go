package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// Create makes a new volume from the generated NewVolume request.
//
// POST /volumes
func (s *Service) Create(ctx context.Context, req api.NewVolume) (*api.Volume, error) {
	resp, err := s.api.PostVolumesWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
