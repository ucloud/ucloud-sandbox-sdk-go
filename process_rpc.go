package sandbox

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

const processServiceName = "process.Process"

type processConfig struct {
	Cmd  string            `json:"cmd"`
	Args []string          `json:"args"`
	Envs map[string]string `json:"envs,omitempty"`
	Cwd  string            `json:"cwd,omitempty"`
}

type startRequest struct {
	Process processConfig `json:"process"`
	Pty     *ptyConfig    `json:"pty,omitempty"`
	Tag     string        `json:"tag,omitempty"`
	Stdin   bool          `json:"stdin,omitempty"`
}

type ptyConfig struct {
	Size *ptySizeMsg `json:"size,omitempty"`
}

type ptySizeMsg struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type connectRequest struct {
	Process processSelector `json:"process"`
}

type processEvent struct {
	Event processEventData `json:"event"`
}

type processEventData struct {
	Start     *processStartEvent `json:"start,omitempty"`
	Data      *processDataEvent  `json:"data,omitempty"`
	End       *processEndEvent   `json:"end,omitempty"`
	Keepalive *struct{}          `json:"keepalive,omitempty"`
}

type processStartEvent struct {
	PID int `json:"pid"`
}

type processDataEvent struct {
	Stdout []byte `json:"stdout,omitempty"`
	Stderr []byte `json:"stderr,omitempty"`
	Pty    []byte `json:"pty,omitempty"`
}

func (d *processDataEvent) UnmarshalJSON(data []byte) error {
	var raw struct {
		Stdout string `json:"stdout,omitempty"`
		Stderr string `json:"stderr,omitempty"`
		Pty    string `json:"pty,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	decode := func(s string) []byte {
		if s == "" {
			return nil
		}
		b, _ := base64.StdEncoding.DecodeString(s)
		return b
	}
	d.Stdout = decode(raw.Stdout)
	d.Stderr = decode(raw.Stderr)
	d.Pty = decode(raw.Pty)
	return nil
}

type processEndEvent struct {
	ExitCode int    `json:"exitCode"`
	Error    string `json:"error,omitempty"`
}

type listProcessesResponse struct {
	Processes []processInfoRaw `json:"processes"`
}

type processInfoRaw struct {
	PID    int           `json:"pid"`
	Tag    string        `json:"tag"`
	Config processConfig `json:"config"`
}

type processSelector struct {
	PID int `json:"pid"`
}

type processInputData struct {
	Stdin string `json:"stdin,omitempty"`
	Pty   string `json:"pty,omitempty"`
}

type sendSignalRequest struct {
	Process processSelector `json:"process"`
	Signal  string          `json:"signal"`
}

type sendInputRequest struct {
	Process processSelector  `json:"process"`
	Input   processInputData `json:"input"`
}

type updateRequest struct {
	Process processSelector `json:"process"`
	Pty     *ptyConfig      `json:"pty,omitempty"`
}

func newSendStdinRequest(pid int, data []byte) sendInputRequest {
	return sendInputRequest{Process: processSelector{PID: pid}, Input: processInputData{Stdin: base64.StdEncoding.EncodeToString(data)}}
}

func newSendPtyRequest(pid int, data []byte) sendInputRequest {
	return sendInputRequest{Process: processSelector{PID: pid}, Input: processInputData{Pty: base64.StdEncoding.EncodeToString(data)}}
}

func isRPCCode(err error, code string) bool {
	var rpcErr *RPCError
	return errors.As(err, &rpcErr) && rpcErr.Code == code
}
