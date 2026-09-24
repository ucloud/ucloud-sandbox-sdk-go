package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// FileUploadLink asks where to upload the file bundle identified by hash.
//
// Used by Build; call it directly only when driving the build sequence
// yourself.
//
// GET /templates/{templateID}/files/{hash}
func (s *Service) FileUploadLink(ctx context.Context, templateID, hash string) (*api.TemplateBuildFileUpload, error) {
	resp, err := s.t.API().GetTemplatesTemplateIDFilesHashWithResponse(ctx, templateID, hash)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
