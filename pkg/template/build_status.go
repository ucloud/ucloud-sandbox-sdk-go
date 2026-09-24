package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// BuildStatus reports where a build has got to, along with any logs produced
// since LogsOffset.
//
// GET /templates/{templateID}/builds/{buildID}/status
func (s *Service) BuildStatus(ctx context.Context, templateID, buildID string, params *api.TemplateBuildStatusParams) (*api.TemplateBuildInfo, error) {
	resp, err := s.t.API().GetTemplatesTemplateIDBuildsBuildIDStatusWithResponse(ctx, templateID, buildID, params)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
