package secret

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// GetInfo returns a secret's metadata, never its value.
//
// secret is either the identifier ("sec_...") or the secret's name.
//
// GET /secrets/{secretID}
func (s *Service) Get(ctx context.Context, secret string) (*api.Secret, error) {
	resp, err := s.t.API().GetSecretsSecretIDWithResponse(ctx, secret)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
