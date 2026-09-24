package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Fork starts copies of a running sandbox.
//
// All forks boot from one snapshot, which is captured once however many are
// asked for. Each fork then succeeds or fails on its own, so the returned
// slice has one entry per requested fork, each carrying either a Sandbox or an
// error. A failure of one fork is not an error from Fork itself.
//
// POST /sandboxes/{sandboxID}/fork
func (s *Service) Fork(ctx context.Context, sandboxID string, req api.SandboxForkRequest) ([]api.SandboxForkResult, error) {
	resp, err := s.t.API().PostSandboxesSandboxIDForkWithResponse(ctx, sandboxID, req)
	if err != nil {
		return nil, err
	}
	forked, err := transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return *forked, nil
}
