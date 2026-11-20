package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// Resize expands a volume to sizeGiB. Shrinking is not supported; requesting
// the current capacity succeeds without changing it.
//
// PATCH /v2/volumes/{volumeID}
func (s *Service) Resize(ctx context.Context, volumeID api.VolumeID, sizeGiB int64) (*api.Volume, error) {
	resp, err := s.api.ResizeVolumeWithResponse(ctx, volumeID, api.ResizeVolume{SizeGiB: sizeGiB})
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
