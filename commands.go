package sandbox

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

type Commands struct {
	sandbox *Sandbox
	rpc     *RPC
}

func newCommands(s *Sandbox, rpc *RPC) *Commands {
	return &Commands{sandbox: s, rpc: rpc}
}

func (c *Commands) Run(ctx context.Context, cmd string, opts ...CommandOption) (*CommandResult, error) {
	cfg := c.applyOpts(opts)
	stream, pid, err := c.startStream(ctx, cmd, cfg)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	return c.consumeStream(stream, pid, cfg, nil, nil, nil)
}

func (c *Commands) Start(ctx context.Context, cmd string, opts ...CommandOption) (*CommandHandle, error) {
	cfg := c.applyOpts(opts)
	stream, pid, err := c.startStream(ctx, cmd, cfg)
	if err != nil {
		return nil, err
	}
	handle := &CommandHandle{PID: pid, done: make(chan struct{}), sandbox: c.sandbox, rpc: c.rpc}
	go func() {
		defer stream.Close()
		result, streamErr := c.consumeStream(stream, pid, cfg, &handle.stdout, &handle.stderr, &handle.mu)
		handle.mu.Lock()
		handle.result = result
		handle.err = streamErr
		handle.mu.Unlock()
		close(handle.done)
	}()
	return handle, nil
}

func (c *Commands) Connect(ctx context.Context, pid int, opts ...CommandOption) (*CommandHandle, error) {
	cfg := c.applyOpts(opts)
	req := connectRequest{Process: processSelector{PID: pid}}
	timeoutMs := 0
	if cfg.timeout > 0 {
		timeoutMs = cfg.timeout * 1000
	}
	stream, err := c.rpc.CallServerStream(ctx, processServiceName, "Connect", req, timeoutMs)
	if err != nil {
		return nil, &SandboxError{Message: fmt.Sprintf("connect process %d: %v", pid, err), Cause: err}
	}
	handle := &CommandHandle{PID: pid, done: make(chan struct{}), sandbox: c.sandbox, rpc: c.rpc}
	go func() {
		defer stream.Close()
		result, streamErr := c.consumeStream(stream, pid, cfg, &handle.stdout, &handle.stderr, &handle.mu)
		handle.mu.Lock()
		handle.result = result
		handle.err = streamErr
		handle.mu.Unlock()
		close(handle.done)
	}()
	return handle, nil
}

func (c *Commands) List(ctx context.Context) ([]ProcessInfo, error) {
	var resp listProcessesResponse
	if err := c.rpc.CallUnary(ctx, processServiceName, "List", struct{}{}, &resp); err != nil {
		return nil, &SandboxError{Message: fmt.Sprintf("list processes: %v", err), Cause: err}
	}
	processes := make([]ProcessInfo, len(resp.Processes))
	for i, p := range resp.Processes {
		processes[i] = ProcessInfo{PID: p.PID, Tag: p.Tag, Cmd: p.Config.Cmd, Args: p.Config.Args, Envs: p.Config.Envs, Cwd: p.Config.Cwd}
	}
	return processes, nil
}

func (c *Commands) Kill(ctx context.Context, pid int) (bool, error) {
	req := sendSignalRequest{Process: processSelector{PID: pid}, Signal: "SIGNAL_SIGKILL"}
	if err := c.rpc.CallUnary(ctx, processServiceName, "SendSignal", req, nil); err != nil {
		if isRPCCode(err, RPCCodeNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Commands) SendStdin(ctx context.Context, pid int, data string) error {
	return c.rpc.CallUnary(ctx, processServiceName, "SendInput", newSendStdinRequest(pid, []byte(data)), nil)
}

func (c *Commands) applyOpts(opts []CommandOption) *commandConfig {
	cfg := &commandConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func (c *Commands) startStream(ctx context.Context, cmd string, cfg *commandConfig) (*RPCStream, int, error) {
	user := c.sandbox.resolveUsername(cfg.user)
	req := startRequest{
		Process: processConfig{Cmd: "/bin/bash", Args: []string{"-l", "-c", cmd}, Envs: cfg.envVars, Cwd: cfg.cwd},
		Stdin:   cfg.stdin && c.sandbox.supportsStdin(),
	}
	timeout := cfg.timeout
	if !cfg.timeoutSet {
		timeout = DefaultCommandTimeout
	}
	timeoutMs := 0
	if timeout > 0 {
		timeoutMs = timeout * 1000
	}
	stream, err := c.rpc.CallServerStream(ctx, processServiceName, "Start", req, timeoutMs, buildAuthHeader(user))
	if err != nil {
		return nil, 0, &SandboxError{Message: fmt.Sprintf("start command: %v", err), Cause: err}
	}
	var firstEvent processEvent
	if err := stream.Next(&firstEvent); err != nil {
		stream.Close()
		return nil, 0, &SandboxError{Message: fmt.Sprintf("read start event: %v", err), Cause: err}
	}
	pid := 0
	if firstEvent.Event.Start != nil {
		pid = firstEvent.Event.Start.PID
	}
	return stream, pid, nil
}

func (c *Commands) consumeStream(stream *RPCStream, pid int, cfg *commandConfig, stdout, stderr *strings.Builder, mu *sync.Mutex) (*CommandResult, error) {
	if stdout == nil {
		stdout = &strings.Builder{}
	}
	if stderr == nil {
		stderr = &strings.Builder{}
	}
	for {
		var event processEvent
		err := stream.Next(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &SandboxError{Message: fmt.Sprintf("stream error: %v", err), Cause: err}
		}
		if event.Event.Data != nil {
			if len(event.Event.Data.Stdout) > 0 {
				s := string(event.Event.Data.Stdout)
				if mu != nil {
					mu.Lock()
				}
				stdout.WriteString(s)
				if mu != nil {
					mu.Unlock()
				}
				if cfg.onStdout != nil {
					cfg.onStdout(s)
				}
			}
			if len(event.Event.Data.Stderr) > 0 {
				s := string(event.Event.Data.Stderr)
				if mu != nil {
					mu.Lock()
				}
				stderr.WriteString(s)
				if mu != nil {
					mu.Unlock()
				}
				if cfg.onStderr != nil {
					cfg.onStderr(s)
				}
			}
		}
		if event.Event.End != nil {
			result := &CommandResult{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: event.Event.End.ExitCode, Error: event.Event.End.Error}
			if result.ExitCode != 0 {
				return result, &CommandExitError{Stdout: result.Stdout, Stderr: result.Stderr, ExitCode: result.ExitCode, Message: result.Error}
			}
			return result, nil
		}
	}
	return &CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

type CommandHandle struct {
	PID     int
	mu      sync.Mutex
	stdout  strings.Builder
	stderr  strings.Builder
	done    chan struct{}
	result  *CommandResult
	err     error
	sandbox *Sandbox
	rpc     *RPC
}

func (h *CommandHandle) Stdout() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.result != nil {
		return h.result.Stdout
	}
	return h.stdout.String()
}

func (h *CommandHandle) Stderr() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.result != nil {
		return h.result.Stderr
	}
	return h.stderr.String()
}

func (h *CommandHandle) Wait(ctx context.Context) (*CommandResult, error) {
	select {
	case <-h.done:
		h.mu.Lock()
		defer h.mu.Unlock()
		return h.result, h.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (h *CommandHandle) Kill(ctx context.Context) error {
	return h.rpc.CallUnary(ctx, processServiceName, "SendSignal", sendSignalRequest{Process: processSelector{PID: h.PID}, Signal: "SIGNAL_SIGKILL"}, nil)
}

func (h *CommandHandle) SendStdin(ctx context.Context, data string) error {
	return h.rpc.CallUnary(ctx, processServiceName, "SendInput", newSendStdinRequest(h.PID, []byte(data)), nil)
}
