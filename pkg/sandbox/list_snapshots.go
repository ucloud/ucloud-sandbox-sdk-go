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
		// Each page is asked for with its own copy of the params, so the
		// cursor moves forward without the caller's struct being written to.
		query := snapshotPageParams(params, token)

		resp, err := s.t.API().GetSnapshotsWithResponse(ctx, &query)
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

// snapshotPageParams copies params for one request of the listing and points
// it at token, the cursor the page before it returned. An empty token leaves
// the caller's own cursor in place, so a listing can be resumed from one.
func snapshotPageParams(params *api.SnapshotListParams, token string) api.SnapshotListParams {
	query := api.SnapshotListParams{}
	if params != nil {
		query = *params
	}
	if token != "" {
		query.NextToken = &token
	}
	return query
}
