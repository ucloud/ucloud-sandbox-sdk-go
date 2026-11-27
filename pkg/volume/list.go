package volume

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

// List returns every volume the team owns.
//
// The endpoint is not paginated, so this returns a slice rather than a
// paginator. The result is never nil.
//
// GET /volumes
func (s *Service) List(ctx context.Context) ([]api.Volume, error) {
	resp, err := s.api.GetVolumesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	page, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return *page, nil
}
