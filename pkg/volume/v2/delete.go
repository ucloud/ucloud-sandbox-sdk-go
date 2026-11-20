package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

// Delete removes the database record and then cleans up the cloud resource.
// An already absent record also succeeds with HTTP 204. Cloud cleanup errors
// are returned even though the database deletion has already been committed.
// The server does not check whether a sandbox is still using the volume.
//
// DELETE /v2/volumes/{volumeID}
func (s *Service) Delete(ctx context.Context, volumeID api.VolumeID) error {
	resp, err := s.api.DeleteVolumeWithResponse(ctx, volumeID)
	if err != nil {
		return err
	}
	return transport.Check(resp.HTTPResponse, resp.Body)
}
