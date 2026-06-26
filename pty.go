package sandbox

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type Pty struct {
	sandbox *Sandbox
	rpc     *RPC
}

func newPty(s *Sandbox, rpc *RPC) *Pty {
	return &Pty{sandbox: s, rpc: rpc}
}

type PtyEvent struct {
	Data []byte
}

type PtyHandle struct {
	PID    int
	events chan PtyEvent
	done   chan struct{}
	rpc    *RPC
}

func (h *PtyHandle) Events() <-chan PtyEvent { return h.events }
func (h *PtyHandle) Wait()                  { <-h.done }

func (h *PtyHandle) Kill(ctx context.Context) (bool, error) {
	if err := h.rpc.CallUnary(ctx, processServiceName, "SendSignal", sendSignalRequest{Process: processSelector{PID: h.PID}, Signal: "SIGNAL_SIGKILL"}, nil); err != nil {
		if isRPCCode(err, RPCCodeNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (h *PtyHandle) SendStdin(ctx context.Context, data []byte) error {
	return h.rpc.CallUnary(ctx, processServiceName, "SendInput", newSendPtyRequest(h.PID, data), nil)
}

func (h *PtyHandle) Resize(ctx context.Context, size PtySize) error {
	return h.rpc.CallUnary(ctx, processServiceName, "Update", updateRequest{Process: processSelector{PID: h.PID}, Pty: &ptyConfig{Size: &ptySizeMsg{Cols: size.Cols, Rows: size.Rows}}}, nil)
}

func (p *Pty) Create(ctx context.Context, size PtySize, opts ...CommandOption) (*PtyHandle, error) {
	cfg := &commandConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	user := p.sandbox.resolveUsername(cfg.user)
	envs := map[string]string{"TERM": "xterm-256color", "LANG": "C.UTF-8", "LC_ALL": "C.UTF-8"}
	for k, v := range cfg.envVars {
		envs[k] = v
	}
	req := startRequest{
		Process: processConfig{Cmd: "/bin/bash", Args: []string{"-i", "-l"}, Envs: envs, Cwd: cfg.cwd},
		Pty:     &ptyConfig{Size: &ptySizeMsg{Cols: size.Cols, Rows: size.Rows}},
	}
	timeoutMs := 0
	if cfg.timeout > 0 {
		timeoutMs = cfg.timeout * 1000
	}
	stream, err := p.rpc.CallServerStream(ctx, processServiceName, "Start", req, timeoutMs, buildAuthHeader(user))
	if err != nil {
		return nil, &SandboxError{Message: fmt.Sprintf("create PTY: %v", err), Cause: err}
	}
	var firstEvent processEvent
	if err := stream.Next(&firstEvent); err != nil {
		stream.Close()
		return nil, &SandboxError{Message: fmt.Sprintf("read PTY start event: %v", err), Cause: err}
	}
	pid := 0
	if firstEvent.Event.Start != nil {
		pid = firstEvent.Event.Start.PID
	}
	events := make(chan PtyEvent, 64)
	done := make(chan struct{})
	handle := &PtyHandle{PID: pid, events: events, done: done, rpc: p.rpc}
	var once sync.Once
	go func() {
		defer stream.Close()
		defer once.Do(func() { close(done) })
		defer close(events)
		for {
			var event processEvent
			if err := stream.Next(&event); err != nil {
				return
			}
			if event.Event.Data != nil && len(event.Event.Data.Pty) > 0 {
				select {
				case events <- PtyEvent{Data: event.Event.Data.Pty}:
				default:
				}
			}
			if event.Event.End != nil {
				return
			}
		}
	}()
	return handle, nil
}

func (p *Pty) Kill(ctx context.Context, pid int) (bool, error) {
	if err := p.rpc.CallUnary(ctx, processServiceName, "SendSignal", sendSignalRequest{Process: processSelector{PID: pid}, Signal: "SIGNAL_SIGKILL"}, nil); err != nil {
		if isRPCCode(err, RPCCodeNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *Pty) Connect(ctx context.Context, pid int, opts ...CommandOption) (*PtyHandle, error) {
	cfg := &commandConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	timeoutMs := 0
	if cfg.timeout > 0 {
		timeoutMs = cfg.timeout * 1000
	}
	stream, err := p.rpc.CallServerStream(ctx, processServiceName, "Connect", connectRequest{Process: processSelector{PID: pid}}, timeoutMs)
	if err != nil {
		return nil, &SandboxError{Message: fmt.Sprintf("connect PTY %d: %v", pid, err), Cause: err}
	}
	var firstEvent processEvent
	if err := stream.Next(&firstEvent); err != nil {
		stream.Close()
		return nil, &SandboxError{Message: fmt.Sprintf("read PTY connect event: %v", err), Cause: err}
	}
	events := make(chan PtyEvent, 64)
	done := make(chan struct{})
	handle := &PtyHandle{PID: pid, events: events, done: done, rpc: p.rpc}
	var once sync.Once
	go func() {
		defer stream.Close()
		defer once.Do(func() { close(done) })
		defer close(events)
		for {
			var event processEvent
			if err := stream.Next(&event); err != nil {
				if err != io.EOF {
					return
				}
				return
			}
			if event.Event.Data != nil && len(event.Event.Data.Pty) > 0 {
				select {
				case events <- PtyEvent{Data: event.Event.Data.Pty}:
				default:
				}
			}
			if event.Event.End != nil {
				return
			}
		}
	}()
	return handle, nil
}

func (p *Pty) Resize(ctx context.Context, pid int, size PtySize) error {
	return p.rpc.CallUnary(ctx, processServiceName, "Update", updateRequest{Process: processSelector{PID: pid}, Pty: &ptyConfig{Size: &ptySizeMsg{Cols: size.Cols, Rows: size.Rows}}}, nil)
}

func (p *Pty) SendStdin(ctx context.Context, pid int, data []byte) error {
	return p.rpc.CallUnary(ctx, processServiceName, "SendInput", newSendPtyRequest(pid, data), nil)
}
