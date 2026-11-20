package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// Get queries cloud details and updates the recorded capacity when it differs.
//
// GET /v2/volumes/{volumeID}
func (s *Service) Get(ctx context.Context, volumeID api.VolumeID) (*api.VolumeDetail, error) {
	resp, err := s.api.GetVolumeWithResponse(ctx, volumeID)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
