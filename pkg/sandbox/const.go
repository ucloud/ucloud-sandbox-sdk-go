package sandbox

import (
	"fmt"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

// Metadata the SDK attaches to every sandbox it creates, recording which
// product opened it. Set it with CreateOptions.ManageBy.
const (
	ManageByMetadataKey = "manageby.sandbox.ucloudai.com"

	ManageByUnknown  = "unknown"
	ManageByDefault  = "ucloud-sandbox-sdk-go"
	ManageBySite     = "site"
	ManageByCodeBox  = "codebox"
	ManageByRagView  = "ragview"
	ManageBySkillLab = "skill-lab"
)

// AllTraffic matches every address, for NetworkConfig's allow and deny lists.
const AllTraffic = "0.0.0.0/0"

// ParseManageBy reads the manage-by marker out of a sandbox's metadata,
// reporting ManageByUnknown for anything it does not recognise.
func ParseManageBy(metadataPtr *api.SandboxMetadata) string {
	if metadataPtr == nil {
		return ManageByUnknown
	}
	metadata := *metadataPtr
	switch value := metadata[ManageByMetadataKey]; value {
	case ManageByDefault, ManageBySite, ManageByCodeBox, ManageByRagView, ManageBySkillLab:
		return value
	default:
		return ManageByUnknown
	}
}

func FilterByManageBy(manager string) string {
	return fmt.Sprintf("%s=%s", ManageByMetadataKey, manager)
}
