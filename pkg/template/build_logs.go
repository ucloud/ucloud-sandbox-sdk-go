package template

import (
	"context"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// BuildLogs returns a build's logs. The result is never nil.
//
// This reads the log store directly, which is what you want for a build that
// has already finished. To follow a build as it runs, use WaitForBuild, whose
// OnLogs callback receives entries as the build produces them.
//
// GET /templates/{templateID}/builds/{buildID}/logs
func (s *Service) BuildLogs(ctx context.Context, templateID, buildID string, params *api.TemplateBuildLogsParams) ([]api.BuildLogEntry, error) {
	resp, err := s.t.API().GetTemplatesTemplateIDBuildsBuildIDLogsWithResponse(ctx, templateID, buildID, params)
	if err != nil {
		return nil, err
	}
	logs, err := transport.Parsed(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return logs.Logs, nil
}
