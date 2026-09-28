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
	return transport.Paginate(params, s.t.API().GetSecretsWithResponse)
}
