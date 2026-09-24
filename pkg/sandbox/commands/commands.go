package commands

import (
	"strings"
	"sync"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/process"
)

// Commands runs and manages processes inside a sandbox. Reach it through
// Sandbox.Commands.
type Commands struct {
	sbx *api.Sandbox

	conn *envd.Connection
}

// Options are the optional arguments to the Commands methods.
type Options struct {
	// Cwd is the working directory. Empty uses the sandbox's default.
	Cwd string

	// EnvVars are added to the command's environment.
	EnvVars map[string]string

	// TimeoutSeconds bounds the command. Zero means
	// DefaultCommandTimeoutSeconds; a negative value means no timeout.
	TimeoutSeconds int

	// Stdin keeps the command's stdin open, so SendStdin can write to it.
	Stdin bool

	// OnStdout and OnStderr receive output as it arrives. Both are called
	// from the stream's goroutine, one call at a time.
	OnStdout func(string)
	OnStderr func(string)
}

// Handle is a running command. Wait for it, feed it input, or kill it.
type Handle struct {
	// PID is the process ID inside the sandbox.
	PID int

	cmds *Commands

	mu     sync.Mutex
	stdout strings.Builder
	stderr strings.Builder

	done   chan struct{}
	result *Result
	err    error
}

// Stdout returns everything written to stdout so far.
func (h *Handle) Stdout() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stdout.String()
}

// Stderr returns everything written to stderr so far.
func (h *Handle) Stderr() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stderr.String()
}

// ProcessInfo describes a process running inside a sandbox.
type ProcessInfo struct {
	PID  int
	Tag  string
	Cmd  string
	Args []string
	Envs map[string]string
	Cwd  string
}

// Result is what a finished command produced.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int

	// Error is envd's description of an abnormal exit, empty when the command
	// simply returned a non-zero code.
	Error string
}

func New(sbx *api.Sandbox, conn *envd.Connection) *Commands {
	return &Commands{
		sbx:  sbx,
		conn: conn,
	}
}

// shellCommand wraps a command line so envd runs it through a shell, which is
// what makes pipes, redirection and globbing work.
func shellCommand(cmd string) *process.ProcessConfig {
	return &process.ProcessConfig{
		Cmd:  "/bin/bash",
		Args: []string{"-l", "-c", cmd},
	}
}

// processConfig builds the process description for a command.
func (c *Commands) processConfig(cmd string, opts Options) *process.ProcessConfig {
	config := shellCommand(cmd)
	config.Envs = opts.EnvVars
	if opts.Cwd != "" {
		config.Cwd = &opts.Cwd
	}
	return config
}

// SelectorForPID names a process by its PID.
func SelectorForPID(pid int) *process.ProcessSelector {
	return &process.ProcessSelector{
		Selector: &process.ProcessSelector_Pid{Pid: uint32(pid)},
	}
}
