package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListV2 returns a paginator over the templates the team can see. No request
// is made until the paginator is walked.
//
// GET /v2/templates
func (s *Service) ListV2(ctx context.Context, params *api.TemplateListParamsV2) *transport.Paginator[api.Template] {
	return transport.Paginate(params, s.t.API().GetV2TemplatesWithResponse)
}
