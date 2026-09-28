package sandbox

import (
	"fmt"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

func GetHost(sbx *api.Sandbox, port int) string {
	if sbx.Domain == nil {
		return ""
	}
	return fmt.Sprintf("%d-%s.%s", port, sbx.SandboxID, *sbx.Domain)
}
