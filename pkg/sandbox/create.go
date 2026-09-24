package sandbox

import (
	"context"
	"maps"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Create starts a sandbox and returns a handle to it.
//
// The caller owns the sandbox's lifetime: it lives for TimeoutSeconds unless
// refreshed, and Kill ends it sooner.
//
// POST /sandboxes
func (s *Service) Create(ctx context.Context, req api.NewSandbox, managedBy string) (*api.Sandbox, error) {
	req.Metadata = withManageBy(req.Metadata, managedBy)

	resp, err := s.t.API().PostSandboxesWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	created, err := transport.Parsed(resp.JSON201, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}

	// A template built by an envd too old to talk to is worse than no sandbox:
	// every later call would fail in a way that points at the wrong thing. Shut
	// it down and say what is actually wrong.
	envdVersion := envd.ParseVersion(created.EnvdVersion)
	if !envdVersion.Supports(envd.VersionMinimum) {
		_, _ = s.Kill(ctx, created.SandboxID)
		return nil, &errdefs.TemplateError{
			Message: "template is too old for this SDK; rebuild it with a current template build"}
	}

	return created, nil
}

// withManageBy adds the manage-by marker to a copy of the caller's metadata,
// leaving their map untouched.
func withManageBy(metadata *api.SandboxMetadata, manageBy string) *api.SandboxMetadata {
	merged := make(api.SandboxMetadata, 0)
	if metadata != nil {
		maps.Copy(merged, *metadata)
	}
	if _, ok := merged[ManageByMetadataKey]; !ok {
		merged[ManageByMetadataKey] = orDefault(manageBy, ManageByDefault)
	}
	return &merged
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
