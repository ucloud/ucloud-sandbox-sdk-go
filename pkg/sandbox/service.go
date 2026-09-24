package sandbox

import (
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/files"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/pty"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Service is the entry point for sandbox operations. Get one from
// client.Client.Sandboxes, or build it directly on a transport client.
//
// It is safe for concurrent use.
type Service struct {
	t *transport.Client

	user string
}

// NewService returns a Service backed by t.
func NewService(t *transport.Client) *Service {
	return &Service{t: t}
}

type EnvdService struct {
	sbx  *api.Sandbox
	conn *envd.Connection
}

func (s *Service) Envd(sbx *api.Sandbox, user string) (*EnvdService, error) {
	conn, err := envd.Connect(s.t, sbx, user)
	if err != nil {
		return nil, err
	}
	return &EnvdService{
		sbx:  sbx,
		conn: conn,
	}, nil
}

func (s *EnvdService) Files() *files.Filesystem {
	return files.New(s.sbx, s.conn)
}

func (s *EnvdService) Commands() *commands.Commands {
	return commands.New(s.sbx, s.conn)
}

func (s *EnvdService) Pty() *pty.Pty {
	return pty.New(s.sbx, s.conn)
}
