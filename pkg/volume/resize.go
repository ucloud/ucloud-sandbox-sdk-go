package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// Resize expands a volume to req.TotalSizeGiB whole GiB, subject to the project
// quota. The new size must exceed the current size. Legacy volumes cannot be
// resized.
//
// POST /volumes/{volumeID}/resize
func (s *Service) Resize(ctx context.Context, volumeID api.VolumeID, req api.ResizeVolumeRequest) (*api.Volume, error) {
	resp, err := s.api.ResizeVolumeWithResponse(ctx, volumeID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
