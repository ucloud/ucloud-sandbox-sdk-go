package files

import (
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
)

type Filesystem struct {
	sbx *api.Sandbox

	conn *envd.Connection
}

func New(sbx *api.Sandbox, conn *envd.Connection) *Filesystem {
	return &Filesystem{
		sbx:  sbx,
		conn: conn,
	}
}
