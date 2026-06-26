package sandbox

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type InstructionType string

const (
	InstructionCopy    InstructionType = "COPY"
	InstructionRun     InstructionType = "RUN"
	InstructionEnv     InstructionType = "ENV"
	InstructionWorkdir InstructionType = "WORKDIR"
	InstructionUser    InstructionType = "USER"
)

type RegistryConfig struct {
	Type               string `json:"type"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string `json:"awsRegion,omitempty"`
	ServiceAccountJSON string `json:"serviceAccountJson,omitempty"`
}

type Instruction struct {
	Type            InstructionType `json:"type"`
	Args            []string        `json:"args"`
	Force           bool            `json:"force,omitempty"`
	ForceUpload     *bool           `json:"forceUpload,omitempty"`
	ResolveSymlinks *bool           `json:"resolveSymlinks,omitempty"`
	FilesHash       string          `json:"filesHash,omitempty"`
}

type ReadyCmd struct {
	cmd string
}

func (r ReadyCmd) String() string { return r.cmd }

func WaitForPort(port int) ReadyCmd {
	return ReadyCmd{cmd: fmt.Sprintf("ss -tuln | grep :%d", port)}
}

func WaitForFile(filename string) ReadyCmd {
	return ReadyCmd{cmd: fmt.Sprintf("[ -f %s ]", filename)}
}

func WaitForTimeout(timeoutMs int) ReadyCmd {
	if timeoutMs < 1000 {
		timeoutMs = 1000
	}
	return ReadyCmd{cmd: fmt.Sprintf("sleep %.3f", float64(timeoutMs)/1000.0)}
}

type registryConfig struct {
	username string
	password string
}

type RegistryOption func(*registryConfig)

func WithRegistryAuth(username, password string) RegistryOption {
	return func(c *registryConfig) {
		c.username = username
		c.password = password
	}
}

func applyRegistryOpts(opts []RegistryOption) *registryConfig {
	cfg := &registryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

type TemplateBuilder struct {
	baseImage       string
	baseTemplate    string
	registryConfig  *RegistryConfig
	startCmd        string
	readyCmd        string
	force           bool
	forceNextLayer  bool
	instructions    []Instruction
	fileContextPath string
	ignorePatterns  []string
	err             error
}

type TemplateBuilderOption func(*TemplateBuilder)

func WithFileContextPath(path string) TemplateBuilderOption {
	return func(t *TemplateBuilder) { t.fileContextPath = path }
}

func NewTemplate(opts ...TemplateBuilderOption) *TemplateBuilder {
	t := &TemplateBuilder{baseTemplate: DefaultTemplate}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *TemplateBuilder) FromImage(image string, opts ...RegistryOption) *TemplateBuilder {
	t.baseImage = image
	t.baseTemplate = ""
	cfg := applyRegistryOpts(opts)
	if cfg.username != "" && cfg.password != "" {
		t.registryConfig = &RegistryConfig{Type: "registry", Username: cfg.username, Password: cfg.password}
	} else {
		t.registryConfig = nil
	}
	if t.forceNextLayer {
		t.force = true
	}
	return t
}

func (t *TemplateBuilder) FromTemplate(template string) *TemplateBuilder {
	t.baseTemplate = template
	t.baseImage = ""
	t.registryConfig = nil
	if t.forceNextLayer {
		t.force = true
	}
	return t
}

func (t *TemplateBuilder) FromBaseImage() *TemplateBuilder {
	return t.FromTemplate(DefaultTemplate)
}

func (t *TemplateBuilder) addInstruction(inst Instruction) {
	inst.Force = t.forceNextLayer
	t.instructions = append(t.instructions, inst)
	t.forceNextLayer = false
}

func (t *TemplateBuilder) RunCmd(command string) *TemplateBuilder {
	t.addInstruction(Instruction{Type: InstructionRun, Args: []string{command}})
	return t
}

func (t *TemplateBuilder) RunCmdAsUser(command, user string) *TemplateBuilder {
	args := []string{command}
	if user != "" {
		args = append(args, user)
	}
	t.addInstruction(Instruction{Type: InstructionRun, Args: args})
	return t
}

func (t *TemplateBuilder) SetEnvs(envs map[string]string) *TemplateBuilder {
	if len(envs) == 0 {
		return t
	}
	keys := make([]string, 0, len(envs))
	for k := range envs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(envs)*2)
	for _, k := range keys {
		args = append(args, k, envs[k])
	}
	t.addInstruction(Instruction{Type: InstructionEnv, Args: args})
	return t
}

func (t *TemplateBuilder) SetWorkdir(workdir string) *TemplateBuilder {
	t.addInstruction(Instruction{Type: InstructionWorkdir, Args: []string{workdir}})
	return t
}

func (t *TemplateBuilder) SetUser(user string) *TemplateBuilder {
	t.addInstruction(Instruction{Type: InstructionUser, Args: []string{user}})
	return t
}

func (t *TemplateBuilder) SkipCache() *TemplateBuilder {
	t.forceNextLayer = true
	return t
}

func (t *TemplateBuilder) SetStartCmd(startCmd string, readyCmd ReadyCmd) *TemplateBuilder {
	t.startCmd = startCmd
	t.readyCmd = readyCmd.cmd
	return t
}

func (t *TemplateBuilder) serialize() templateData {
	data := templateData{
		Steps: t.instructions,
		Force: t.force,
	}
	if t.baseImage != "" {
		data.FromImage = t.baseImage
	}
	if t.baseTemplate != "" {
		data.FromTemplate = t.baseTemplate
	}
	if t.registryConfig != nil {
		data.FromImageRegistry = t.registryConfig
	}
	if t.startCmd != "" {
		data.StartCmd = t.startCmd
	}
	if t.readyCmd != "" {
		data.ReadyCmd = t.readyCmd
	}
	return data
}

func (t *TemplateBuilder) ToJSON() (string, error) {
	data := t.serialize()
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *TemplateBuilder) ToDockerfile() (string, error) {
	if t.baseTemplate != "" {
		return "", fmt.Errorf("cannot convert template built from another template to Dockerfile")
	}
	if t.baseImage == "" {
		return "", fmt.Errorf("no base image specified for template")
	}

	var buf strings.Builder
	buf.WriteString("FROM ")
	buf.WriteString(t.baseImage)
	buf.WriteString("\n")

	for _, inst := range t.instructions {
		switch inst.Type {
		case InstructionRun:
			if len(inst.Args) == 0 {
				continue
			}
			if len(inst.Args) > 1 && inst.Args[1] != "" {
				buf.WriteString("USER ")
				buf.WriteString(inst.Args[1])
				buf.WriteString("\n")
			}
			buf.WriteString("RUN ")
			buf.WriteString(inst.Args[0])
			buf.WriteString("\n")
		case InstructionCopy:
			if len(inst.Args) >= 2 {
				buf.WriteString("COPY ")
				buf.WriteString(inst.Args[0])
				buf.WriteString(" ")
				buf.WriteString(inst.Args[1])
				buf.WriteString("\n")
			}
		case InstructionEnv:
			var pairs []string
			for i := 0; i+1 < len(inst.Args); i += 2 {
				pairs = append(pairs, inst.Args[i]+"="+inst.Args[i+1])
			}
			if len(pairs) > 0 {
				buf.WriteString("ENV ")
				buf.WriteString(strings.Join(pairs, " "))
				buf.WriteString("\n")
			}
		case InstructionWorkdir:
			if len(inst.Args) > 0 {
				buf.WriteString("WORKDIR ")
				buf.WriteString(inst.Args[0])
				buf.WriteString("\n")
			}
		case InstructionUser:
			if len(inst.Args) > 0 {
				buf.WriteString("USER ")
				buf.WriteString(inst.Args[0])
				buf.WriteString("\n")
			}
		default:
			buf.WriteString(string(inst.Type))
			buf.WriteString(" ")
			buf.WriteString(strings.Join(inst.Args, " "))
			buf.WriteString("\n")
		}
	}

	if t.startCmd != "" {
		buf.WriteString("ENTRYPOINT ")
		buf.WriteString(t.startCmd)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
