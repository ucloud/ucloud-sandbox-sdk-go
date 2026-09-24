package sandbox

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// LogsV2 returns a sandbox's logs. The result is never nil.
//
// GET /v2/sandboxes/{sandboxID}/logs
func (s *Service) LogsV2(ctx context.Context, sandboxID string, params *api.SandboxLogsParamsV2) ([]api.SandboxLogEntry, error) {
	resp, err := s.t.API().GetV2SandboxesSandboxIDLogsWithResponse(ctx, sandboxID, params)
	if err != nil {
		return nil, err
	}
	logs, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return logs.Logs, nil
}
