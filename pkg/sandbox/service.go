package sandbox

import (
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
