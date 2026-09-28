package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListV2 returns every template the team can see.
//
// GET /v2/templates
func (s *Service) ListV2(ctx context.Context, params *api.TemplateListParamsV2) *transport.Paginator[api.Template] {
	return transport.NewPaginator(func(ctx context.Context, token string) ([]api.Template, string, error) {
		query := templatePageParams(params, token)

		resp, err := s.t.API().GetV2TemplatesWithResponse(ctx, &query)
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

func templatePageParams(params *api.TemplateListParamsV2, token string) api.TemplateListParamsV2 {
	query := api.TemplateListParamsV2{}
	if params != nil {
		query = *params
	}
	if token != "" {
		query.NextToken = &token
	}
	return query
}
