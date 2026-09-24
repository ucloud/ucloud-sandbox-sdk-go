package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListV2 returns a paginator over the team's sandboxes. No request is made
// until the paginator is walked.
// GET /v2/sandboxes
func (s *Service) ListV2(ctx context.Context, params *api.SandboxListParamsV2) *transport.Paginator[api.ListedSandbox] {
	return transport.NewPaginator(func(ctx context.Context, token string) ([]api.ListedSandbox, string, error) {
		// Each page is asked for with its own copy of the params, so the
		// cursor moves forward without the caller's struct being written to.
		query := sandboxPageParams(params, token)

		resp, err := s.t.API().GetV2SandboxesWithResponse(ctx, &query)
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

// sandboxPageParams copies params for one request of the listing and points it
// at token, the cursor the page before it returned. An empty token leaves the
// caller's own cursor in place, so a listing can be resumed from one.
func sandboxPageParams(params *api.SandboxListParamsV2, token string) api.SandboxListParamsV2 {
	query := api.SandboxListParamsV2{}
	if params != nil {
		query = *params
	}
	if token != "" {
		query.NextToken = &token
	}
	return query
}
