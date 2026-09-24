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
		resp, err := s.t.API().GetV2SandboxesWithResponse(ctx, params)
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
