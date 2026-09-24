package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Get returns a sandbox's current state.
//
// GET /sandboxes/{sandboxID}
func (s *Service) Get(ctx context.Context, sandboxID string) (*api.SandboxDetail, error) {
	resp, err := s.t.API().GetSandboxesSandboxIDWithResponse(ctx, sandboxID)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
}
