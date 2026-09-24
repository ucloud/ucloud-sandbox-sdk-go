package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// StartBuildV2 starts a build that CreateV3 allocated, using the template the
// builder describes.
//
// POST /v2/templates/{templateID}/builds/{buildID}
func (s *Service) StartBuildV2(ctx context.Context, templateID, buildID string, req api.TemplateBuildStartV2) error {
	resp, err := s.t.API().PostV2TemplatesTemplateIDBuildsBuildIDWithResponse(ctx, templateID, buildID, req)
	if err != nil {
		return err
	}
	return transport.Check(resp.HTTPResponse, resp.Body)
}
