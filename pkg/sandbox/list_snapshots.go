package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// ListSnapshots returns a paginator over the team's snapshots. No request is
// made until the paginator is walked.
//
// GET /snapshots
func (s *Service) ListSnapshots(ctx context.Context, params *api.SnapshotListParams) *transport.Paginator[api.SnapshotInfo] {
	return transport.Paginate(params, s.t.API().GetSnapshotsWithResponse)
}
