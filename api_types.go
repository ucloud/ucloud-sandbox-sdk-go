package sandbox

// Internal API request/response wire types.

type autoResumeObject struct {
	Enabled bool `json:"enabled"`
}

func toAutoResumeObject(p *AutoResumePolicy) *autoResumeObject {
	if p == nil {
		return nil
	}
	return &autoResumeObject{Enabled: *p == AutoResumePolicyOn}
}

type snapshotRequest struct {
	Name string `json:"name,omitempty"`
}

type snapshotResponse struct {
	SnapshotID string `json:"snapshotID"`
}

type createSandboxResponse struct {
	SandboxID          string `json:"sandboxID"`
	ClientID           string `json:"clientID"`
	Domain             string `json:"domain"`
	EnvdVersion        string `json:"envdVersion"`
	EnvdAccessToken    string `json:"envdAccessToken"`
	TrafficAccessToken string `json:"trafficAccessToken"`
}

type connectSandboxResponse struct {
	EnvdVersion        string `json:"envdVersion"`
	EnvdAccessToken    string `json:"envdAccessToken"`
	TrafficAccessToken string `json:"trafficAccessToken"`
	Domain             string `json:"domain"`
}

type createSandboxRequest struct {
	TemplateID          string            `json:"templateID"`
	Timeout             int               `json:"timeout,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	EnvVars             map[string]string `json:"envVars,omitempty"`
	AutoPause           *bool             `json:"autoPause,omitempty"`
	Secure              *bool             `json:"secure,omitempty"`
	AllowInternetAccess *bool             `json:"allowInternetAccess,omitempty"`
	Network             *NetworkOpts      `json:"network,omitempty"`
	AutoResume          *autoResumeObject `json:"autoResume,omitempty"`
	VolumeMounts        []VolumeMount     `json:"volumeMounts,omitempty"`
	MCP                 MCPConfig         `json:"mcp,omitempty"`
}

type templateUpdateRequest struct {
	Public bool `json:"public"`
}

type templateUpdateResponse struct {
	Names []string `json:"names"`
}

type templateFileUploadLink struct {
	Present   bool   `json:"present"`
	URL       string `json:"url"`
	UploadURL string `json:"uploadUrl"`
}

type templateData struct {
	FromImage         string          `json:"fromImage,omitempty"`
	FromTemplate      string          `json:"fromTemplate,omitempty"`
	FromImageRegistry *RegistryConfig `json:"fromImageRegistry,omitempty"`
	StartCmd          string          `json:"startCmd,omitempty"`
	ReadyCmd          string          `json:"readyCmd,omitempty"`
	Steps             []Instruction   `json:"steps"`
	Force             bool            `json:"force"`
}

type requestBuildRequest struct {
	Name     string   `json:"name"`
	Tags     []string `json:"tags,omitempty"`
	CPUCount int      `json:"cpuCount"`
	MemoryMB int      `json:"memoryMB"`
}

type requestBuildResponse struct {
	TemplateID string   `json:"templateID"`
	BuildID    string   `json:"buildID"`
	Tags       []string `json:"tags,omitempty"`
}

type buildStatusAPIResponse struct {
	TemplateID string              `json:"templateID"`
	BuildID    string              `json:"buildID"`
	Status     TemplateBuildStatus `json:"status"`
	Logs       []string            `json:"logs"`
	LogEntries []LogEntry          `json:"logEntries,omitempty"`
	Reason     *BuildStatusReason  `json:"reason,omitempty"`
}
