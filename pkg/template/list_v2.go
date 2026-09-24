package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListV2 returns every template the team can see. The endpoint is not
// paginated, so this returns a slice. The result is never nil.
//
// GET /v2/templates
func (s *Service) ListV2(ctx context.Context) ([]api.Template, error) {
	resp, err := s.t.API().GetV2TemplatesWithResponse(ctx, &api.GetV2TemplatesParams{})
	if err != nil {
		return nil, err
	}
	page, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return *page, nil
}
