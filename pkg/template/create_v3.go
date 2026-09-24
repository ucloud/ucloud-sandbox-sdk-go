package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// CreateV3 registers a template and allocates a build for it. It does not run
// the build; see Build, which does the whole sequence.
//
// name may carry a tag, as in "my-template:v1", which counts as if the tag had
// been passed in CreateV3Options.Tags.
//
// POST /v3/templates
func (s *Service) CreateV3(ctx context.Context, req api.TemplateBuildRequestV3) (*api.TemplateRequestResponseV3, error) {
	resp, err := s.t.API().PostV3TemplatesWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON202, resp.HTTPResponse, resp.Body)
}
