package sandbox

import (
	"context"
	"fmt"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/files"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/pty"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

type Sandbox struct {
	api.Sandbox

	envdConn *envd.Connection
	client   *transport.Client
}

func wrap(client *transport.Client, sbx *api.Sandbox) (*Sandbox, error) {
	if sbx == nil {
		return nil, errdefs.ErrNotFound
	}
	conn, err := envd.Connect(client, sbx)
	if err != nil {
		return nil, err
	}
	return &Sandbox{
		Sandbox:  *sbx,
		envdConn: conn,
		client:   client,
	}, nil
}

func (s *Sandbox) Commands() *commands.Commands {
	return commands.New(&s.Sandbox, s.envdConn)
}

func (s *Sandbox) Files() *files.Filesystem {
	return files.New(s.envdConn)
}

func (s *Sandbox) Pty() *pty.Pty {
	return pty.New(s.envdConn)
}

func (s *Sandbox) GetHost(port int) string {
	var domain string
	if s.Domain != nil {
		domain = *s.Domain
	} else {
		domain = s.client.Domain()
	}
	return fmt.Sprintf("%d-%s.%s", port, s.SandboxID, domain)
}

func (s *Sandbox) Kill(ctx context.Context) (bool, error) {
	return NewService(s.client).Kill(ctx, s.SandboxID)
}

func (s *Sandbox) Pause(ctx context.Context, req api.SandboxPauseRequest) error {
	return NewService(s.client).Pause(ctx, s.SandboxID, req)
}

func (s *Sandbox) Fork(ctx context.Context, req api.SandboxForkRequest) ([]api.SandboxForkResult, error) {
	return NewService(s.client).Fork(ctx, s.SandboxID, req)
}

func (s *Sandbox) LogsV2(ctx context.Context, params *api.SandboxLogsParamsV2) ([]api.SandboxLogEntry, error) {
	return NewService(s.client).LogsV2(ctx, s.SandboxID, params)
}

func (s *Sandbox) Metrics(ctx context.Context, params *api.SandboxMetricsParams) ([]api.SandboxMetric, error) {
	return NewService(s.client).Metrics(ctx, s.SandboxID, params)
}

func (s *Sandbox) Refresh(ctx context.Context, durationSeconds int) error {
	return NewService(s.client).Refresh(ctx, s.SandboxID, durationSeconds)
}

func (s *Sandbox) SetTimeout(ctx context.Context, sandboxID string, timeoutSeconds int) error {
	return NewService(s.client).SetTimeout(ctx, s.SandboxID, timeoutSeconds)
}

func (s *Sandbox) UpdateNetwork(ctx context.Context, update api.SandboxNetworkUpdateConfig) error {
	return NewService(s.client).UpdateNetwork(ctx, s.SandboxID, update)
}

func (s *Sandbox) GetDetail(ctx context.Context) (*api.SandboxDetail, error) {
	return NewService(s.client).Get(ctx, s.SandboxID)
}
