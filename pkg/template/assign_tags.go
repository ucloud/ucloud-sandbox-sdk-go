package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// AssignTags points tags at a build.
//
// target names the build, as a template name optionally carrying a tag, for
// example "my-template" or "my-template:v1".
//
// POST /templates/tags
func (s *Service) AssignTags(ctx context.Context, req api.AssignTemplateTagsRequest) (*api.AssignedTemplateTags, error) {
	resp, err := s.t.API().PostTemplatesTagsWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
