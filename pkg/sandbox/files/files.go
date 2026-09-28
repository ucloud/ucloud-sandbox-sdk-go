package files

import (
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
)

type Filesystem struct {
	conn *envd.Connection
	user string
}

func New(conn *envd.Connection) *Filesystem {
	return &Filesystem{
		conn: conn,
	}
}

func (f *Filesystem) User(user string) *Filesystem {
	f.user = user
	return f
}
