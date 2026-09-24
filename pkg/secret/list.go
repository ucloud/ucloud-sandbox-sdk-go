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
		// Each page is asked for with its own copy of the params, so the
		// cursor moves forward without the caller's struct being written to.
		query := pageParams(params, token)

		resp, err := s.t.API().GetSecretsWithResponse(ctx, &query)
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

// pageParams copies params for one request of the listing and points it at
// token, the cursor the page before it returned. An empty token leaves the
// caller's own cursor in place, so a listing can be resumed from one.
func pageParams(params *api.SecretListParams, token string) api.SecretListParams {
	query := api.SecretListParams{}
	if params != nil {
		query = *params
	}
	if token != "" {
		query.NextToken = &token
	}
	return query
}
