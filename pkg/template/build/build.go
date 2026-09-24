package build

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// Polling bounds for WaitForBuild. The interval starts short so a cached build
// is noticed almost at once, then backs off so a long build does not hammer the
// control plane.
const (
	initialPollInterval = 200 * time.Millisecond
	maxPollInterval     = 2 * time.Second
)

type Logger func(api.BuildLogEntry)

type Builder struct {
	req    api.TemplateBuildStartV2
	logger Logger

	baseImage bool

	fileContextPath string

	err error
}

type StepType string

const (
	StepFrom       StepType = "FROM"
	StepCopy       StepType = "COPY"
	StepAdd        StepType = "ADD"
	StepRun        StepType = "RUN"
	StepEnv        StepType = "ENV"
	StepArg        StepType = "ARG"
	StepWorkdir    StepType = "WORKDIR"
	StepUser       StepType = "USER"
	StepEntrypoint StepType = "ENTRYPOINT"
	StepCmd        StepType = "CMD"
)

// FromImage starts the template from a public container image.
func FromImage(image string) *Builder {
	return &Builder{
		req: api.TemplateBuildStartV2{
			FromImage: &image,
		},
	}
}

// FromTemplate starts the template from another template rather than an image.
func FromTemplate(template string) *Builder {
	return &Builder{
		req: api.TemplateBuildStartV2{
			FromTemplate: &template,
		},
	}
}

// FromBaseImage starts the template from the platform's base image for the
// builder's region.
func FromBaseImage() *Builder {
	return &Builder{
		req: api.TemplateBuildStartV2{},

		baseImage: true,
	}
}

func (b *Builder) SetImageRegistryAuth(username, password string) *Builder {
	registry := &api.FromImageRegistry{}
	registry.FromGeneralRegistry(api.GeneralRegistry{
		Type:     api.Registry,
		Username: username,
		Password: password,
	})
	b.req.FromImageRegistry = registry
	return b
}

func (b *Builder) Force(f bool) *Builder {
	b.req.Force = &f
	return b
}

type RunCmdOption func(step *api.TemplateStep)

func RunCmdWithUser(user string) RunCmdOption {
	return func(step *api.TemplateStep) {
		args := *step.Args
		if user != "" {
			args = append(args, user)
		}
		step.Args = &args
	}
}

func RunCmdSkipCache() RunCmdOption {
	return func(step *api.TemplateStep) {
		step.Force = new(true)
	}
}

// RunCmd runs a shell command during the build.
func (b *Builder) RunCmd(command string, opts ...RunCmdOption) *Builder {
	step := api.TemplateStep{
		Type: string(StepRun),
		Args: new([]string{command}),
	}
	for _, opt := range opts {
		opt(&step)
	}
	return b.AddStep(step)
}

// SetEnvs sets environment variables for the remaining steps and for sandboxes
// started from the template.
//
// The keys are sorted, so the same map always produces the same step and
// therefore the same cache key.
func (b *Builder) SetEnvs(envs map[string]string) *Builder {
	if len(envs) == 0 {
		return b
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

	return b.AddStep(api.TemplateStep{
		Type: string(StepEnv),
		Args: &args,
	})
}

// SetWorkdir sets the working directory for the remaining steps.
func (b *Builder) SetWorkdir(workdir string) *Builder {
	return b.AddStep(api.TemplateStep{
		Type: string(StepWorkdir),
		Args: new([]string{workdir}),
	})
}

// SetUser sets the user for the remaining steps.
func (b *Builder) SetUser(user string) *Builder {
	return b.AddStep(api.TemplateStep{
		Type: string(StepUser),
		Args: new([]string{user}),
	})
}

// Copy copies files from the local file context into the template.
//
// src is relative to Builder.SetFileContextPath and must not escape it.
// The files are hashed and uploaded by Build before the build starts.
func (b *Builder) Copy(src, dest string) *Builder {
	if b.err != nil {
		return b
	}
	if b.fileContextPath == "" {
		b.err = fmt.Errorf(
			"template: COPY %q needs a file context; call Builder.SetFileContextPath",
			src)
		return b
	}

	hash, err := hashFileContext(src, dest, b.fileContextPath)
	if err != nil {
		b.err = err
		return b
	}

	return b.AddStep(api.TemplateStep{
		Type:      string(StepCopy),
		FilesHash: &hash,
		Args:      new([]string{src, dest}),
	})
}

func (b *Builder) AddStep(step api.TemplateStep) *Builder {
	var steps []api.TemplateStep
	if b.req.Steps == nil {
		steps = make([]api.TemplateStep, 0, 1)
	} else {
		steps = *b.req.Steps
	}
	steps = append(steps, step)
	b.req.Steps = &steps
	return b
}

func (b *Builder) SetStartCmd(startCmd string) *Builder {
	b.req.StartCmd = &startCmd
	return b
}

func (b *Builder) SetReadyCmd(readyCmd string) *Builder {
	b.req.ReadyCmd = &readyCmd
	return b
}

func (b *Builder) WaitForPortReady(port int) *Builder {
	return b.SetReadyCmd(fmt.Sprintf("ss -tuln | grep :%d", port))
}

func (b *Builder) WaitForFileReady(filename string) *Builder {
	return b.SetReadyCmd(fmt.Sprintf("[ -f %s ]", filename))
}

func (b *Builder) WaitForTimeout(d time.Duration) *Builder {
	timeoutMs := max(d.Milliseconds(), 1000)
	return b.SetReadyCmd(fmt.Sprintf("sleep %.3f", float64(timeoutMs)/1000.0))
}

func (b *Builder) SetFileContextPath(path string) *Builder {
	b.fileContextPath = path
	return b
}

func (b *Builder) SetLogger(logger Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) Build(ctx context.Context, c *client.Client, template api.TemplateBuildRequestV3) (*api.TemplateRequestResponseV3, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.baseImage {
		b.req.FromImage = new(GetBaseImage(c.Region()))
	}

	b.emit("Requesting build for template")
	info, err := c.Templates().CreateV3(ctx, template)
	if err != nil {
		return nil, err
	}

	if err := b.uploadFiles(ctx, c, info.TemplateID); err != nil {
		return nil, err
	}

	b.emit("Template created with ID %s, build %s", info.TemplateID, info.BuildID)
	b.emit("Starting build...")

	if err := c.Templates().StartBuildV2(ctx, info.TemplateID, info.BuildID, b.req); err != nil {
		return nil, err
	}

	if _, err := b.waitForBuild(ctx, c, info.TemplateID, info.BuildID); err != nil {
		return info, err
	}

	return info, nil
}

func (b *Builder) waitForBuild(ctx context.Context, c *client.Client, templateID, buildID string) (*api.TemplateBuildStatus, error) {
	var logsOffset int32
	interval := initialPollInterval

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		status, err := c.Templates().BuildStatus(ctx, templateID, buildID, &api.TemplateBuildStatusParams{
			LogsOffset: &logsOffset,
		})
		if err != nil {
			return nil, err
		}

		if b.logger != nil {
			for _, entry := range status.LogEntries {
				b.logger(entry)
			}
		}
		logsOffset += int32(len(status.LogEntries))

		switch status.Status {
		case api.TemplateBuildStatusReady:
			return &status.Status, nil

		case api.TemplateBuildStatusError:
			message := "template build failed"
			if status.Reason != nil && status.Reason.Message != "" {
				message = status.Reason.Message
			}
			return &status.Status, &errdefs.BuildError{
				SandboxError: errdefs.SandboxError{Message: message},
				BuildID:      buildID,
				TemplateID:   templateID,
			}

		case api.TemplateBuildStatusBuilding, api.TemplateBuildStatusWaiting:
			// Still going.

		default:
			return &status.Status, &errdefs.BuildError{
				SandboxError: errdefs.SandboxError{
					Message: fmt.Sprintf("unknown build status %q", status.Status),
				},
				BuildID:    buildID,
				TemplateID: templateID,
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		interval = min(interval*2, maxPollInterval)
	}
}

func (b *Builder) emit(format string, args ...any) {
	if b.logger == nil {
		return
	}
	b.logger(api.BuildLogEntry{
		Timestamp: time.Now(),
		Level:     api.LogLevelInfo,
		Message:   fmt.Sprintf(format, args...),
	})
}
