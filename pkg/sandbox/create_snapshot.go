package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// CreateSnapshot captures a sandbox as a template that can be booted later.
//
// POST /sandboxes/{sandboxID}/snapshots
func (s *Service) CreateSnapshot(ctx context.Context, sandboxID string, req api.SandboxSnapshotRequest) (*api.SnapshotInfo, error) {
	resp, err := s.t.API().PostSandboxesSandboxIDSnapshotsWithResponse(ctx, sandboxID, req)
	if err != nil {
		return nil, err
	}
	return transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
}
