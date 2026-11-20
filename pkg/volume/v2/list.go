package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// List returns one page of project volumes, ordered by creation time descending
// and then ID descending. params may be nil to use the server defaults.
//
// GET /v2/volumes
func (s *Service) List(ctx context.Context, params *api.ListVolumesParams) (*api.VolumeList, error) {
	resp, err := s.api.ListVolumesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
