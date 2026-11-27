package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// Get returns volume information, including capacity and usage when available.
// Legacy volumes omit the capacity fields and may include a content token.
//
// GET /volumes/{volumeID}
func (s *Service) Get(ctx context.Context, volumeID api.VolumeID) (*api.Volume, error) {
	resp, err := s.api.GetVolumesVolumeIDWithResponse(ctx, volumeID)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
