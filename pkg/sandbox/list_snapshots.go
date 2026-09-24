package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListSnapshots returns a paginator over the team's snapshots. No request is
// made until the paginator is walked.
//
// GET /snapshots
func (s *Service) ListSnapshots(ctx context.Context, params *api.SnapshotListParams) *transport.Paginator[api.SnapshotInfo] {
	return transport.NewPaginator(func(ctx context.Context, token string) ([]api.SnapshotInfo, string, error) {
		resp, err := s.t.API().GetSnapshotsWithResponse(ctx, params)
		if err != nil {
			return nil, "", err
		}
		page, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
		if err != nil {
			return nil, "", err
		}
		return *page, transport.NextTokenFrom(resp.HTTPResponse.Header), nil
	})
}
