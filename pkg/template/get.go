package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Get returns a template together with one page of its builds.
//
// GET /templates/{templateID}
func (s *Service) Get(ctx context.Context, templateID string, params *api.TemplateGetParams) (*api.TemplateWithBuilds, error) {
	resp, err := s.t.API().GetTemplatesTemplateIDWithResponse(ctx, templateID, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
