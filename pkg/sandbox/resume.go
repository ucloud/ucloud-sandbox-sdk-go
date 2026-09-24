package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Resume restarts a paused sandbox and returns a fresh handle to it.
//
// POST /sandboxes/{sandboxID}/resume
func (s *Service) Resume(ctx context.Context, sandboxID string, req api.ResumedSandbox) (*api.Sandbox, error) {
	resp, err := s.t.API().PostSandboxesSandboxIDResumeWithResponse(ctx, sandboxID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
