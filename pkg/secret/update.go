package secret

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Update stores value as the secret's next version and makes it the one served
// to readers that do not name a version. Earlier versions are kept.
//
// secret is either the identifier ("sec_...") or the secret's name. value is
// write-only.
//
// POST /secrets/{secretID}
func (s *Service) Update(ctx context.Context, secret string, req api.SecretUpdate) (*api.Secret, error) {
	resp, err := s.t.API().PostSecretsSecretIDWithResponse(ctx, secret, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
