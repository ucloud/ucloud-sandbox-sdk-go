package sandbox

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
func ParseManageBy(metadata map[string]string) string {
	switch value := metadata[ManageByMetadataKey]; value {
	case ManageByDefault, ManageBySite, ManageByCodeBox, ManageByRagView, ManageBySkillLab:
		return value
	default:
		return ManageByUnknown
	}
}
