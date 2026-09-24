package secret

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Create stores value as the first version of a new secret.
//
// POST /secrets
func (s *Service) Create(ctx context.Context, req api.NewSecret) (*api.Secret, error) {
	resp, err := s.t.API().PostSecretsWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
