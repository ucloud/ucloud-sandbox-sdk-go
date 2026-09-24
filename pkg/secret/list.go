package secret

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// List returns a paginator over the project's secrets, newest cursor first. No
// request is made until the paginator is walked.
//
// GET /secrets
func (s *Service) List(ctx context.Context, params *api.SecretListParams) *transport.Paginator[api.Secret] {
	return transport.NewPaginator(func(ctx context.Context, token string) ([]api.Secret, string, error) {
		resp, err := s.t.API().GetSecretsWithResponse(ctx, params)
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
