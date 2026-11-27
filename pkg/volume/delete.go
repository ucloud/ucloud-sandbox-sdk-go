package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// Delete removes a volume and everything in it. A volume that does not exist is
// reported as false rather than as an error, so deleting twice is not a
// failure.
//
// DELETE /volumes/{volumeID}
func (s *Service) Delete(ctx context.Context, volumeID api.VolumeID) (bool, error) {
	resp, err := s.api.DeleteVolumesVolumeIDWithResponse(ctx, volumeID)
	if err != nil {
		return false, err
	}
	if transport.IsNotFound(resp.HTTPResponse) {
		return false, nil
	}
	if err := transport.Check(resp.HTTPResponse, resp.Body); err != nil {
		return false, err
	}
	return true, nil
}
