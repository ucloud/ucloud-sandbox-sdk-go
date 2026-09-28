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
	return transport.Paginate(params, s.t.API().GetV2SandboxesWithResponse)
}
