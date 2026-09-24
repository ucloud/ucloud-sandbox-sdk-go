package template

import "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"

// Service is the entry point for template operations. Get one from
// client.Client.Templates, or build it directly on a transport client.
//
// It is safe for concurrent use.
type Service struct {
	t *transport.Client
}

// NewService returns a Service backed by t.
func NewService(t *transport.Client) *Service {
	return &Service{t: t}
}
