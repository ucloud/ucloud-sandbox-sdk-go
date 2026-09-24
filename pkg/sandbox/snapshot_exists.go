package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

// SnapshotExists reports whether a snapshot is still on the platform.
//
// snapshot is either the snapshot's ID or its name, optionally tag-qualified.
//
// This asks the listing to filter, so it is one request whatever the team's
// snapshot count. Walking every page and comparing IDs, as earlier versions of
// this SDK did, cost a request per page.
func (s *Service) SnapshotExists(ctx context.Context, snapshot string) (bool, error) {
	page, err := s.ListSnapshots(ctx, &api.SnapshotListParams{
		Name:  &snapshot,
		Limit: new(api.PaginationLimit(1)),
	}).NextItems(ctx)
	if err != nil {
		return false, err
	}
	return len(page) > 0, nil
}
