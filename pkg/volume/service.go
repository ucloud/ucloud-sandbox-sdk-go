package volume

import (
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
)

// Service is the entry point for volume operations. Get one from
// client.Client.Volumes, or build it directly on a transport client.
//
// It is safe for concurrent use.
type Service struct {
	t *transport.Client
}

// NewService returns a Service backed by t.
func NewService(t *transport.Client) *Service {
	return &Service{t: t}
}
