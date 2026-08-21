package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SandboxLogEntry is a single structured entry of a sandbox log.
type SandboxLogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Level     string            `json:"level"`
	Fields    map[string]string `json:"fields"`
}

func (e SandboxLogEntry) String() string { return e.Message }

type sandboxLogsConfig struct {
	cursorMs  *int64
	limit     *int
	direction LogsDirection
	level     string
	search    string
}

// SandboxLogsOption customizes a GetSandboxLogs call.
type SandboxLogsOption func(*sandboxLogsConfig)

// WithSandboxLogsCursor sets the starting timestamp of the logs, in milliseconds.
func WithSandboxLogsCursor(cursorMs int64) SandboxLogsOption {
	return func(c *sandboxLogsConfig) { c.cursorMs = &cursorMs }
}

// WithSandboxLogsLimit caps the number of returned entries (0-1000, defaults to 1000 server-side).
func WithSandboxLogsLimit(limit int) SandboxLogsOption {
	return func(c *sandboxLogsConfig) { c.limit = &limit }
}

// WithSandboxLogsDirection reads logs forward or backward from the cursor.
func WithSandboxLogsDirection(direction LogsDirection) SandboxLogsOption {
	return func(c *sandboxLogsConfig) { c.direction = direction }
}

// WithSandboxLogsLevel excludes entries below the given severity (debug, info, warn, error).
func WithSandboxLogsLevel(level string) SandboxLogsOption {
	return func(c *sandboxLogsConfig) { c.level = level }
}

// WithSandboxLogsSearch filters entries by a case-sensitive substring of the message (max 256 chars).
func WithSandboxLogsSearch(search string) SandboxLogsOption {
	return func(c *sandboxLogsConfig) { c.search = search }
}

type sandboxLogsResponse struct {
	Logs []SandboxLogEntry `json:"logs"`
}

// GetSandboxLogs returns the structured logs of a sandbox.
func (c *Client) GetSandboxLogs(ctx context.Context, sandboxID string, opts ...SandboxLogsOption) ([]SandboxLogEntry, error) {
	if sandboxID == "" {
		return nil, fmt.Errorf("sandboxID is required")
	}

	cfg := &sandboxLogsConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	query := url.Values{}
	if cfg.cursorMs != nil {
		query.Set("cursor", strconv.FormatInt(*cfg.cursorMs, 10))
	}
	if cfg.limit != nil {
		query.Set("limit", strconv.Itoa(*cfg.limit))
	}
	if cfg.direction != "" {
		query.Set("direction", string(cfg.direction))
	}
	if cfg.level != "" {
		query.Set("level", cfg.level)
	}
	if cfg.search != "" {
		query.Set("search", cfg.search)
	}

	path := "/v2/sandboxes/" + url.PathEscape(sandboxID) + "/logs"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp sandboxLogsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("get sandbox logs %s: %w", sandboxID, err)
	}
	if resp.Logs == nil {
		return []SandboxLogEntry{}, nil
	}
	return resp.Logs, nil
}

// GetLogs returns the structured logs of this sandbox.
func (s *Sandbox) GetLogs(ctx context.Context, opts ...SandboxLogsOption) ([]SandboxLogEntry, error) {
	return s.client.GetSandboxLogs(ctx, s.ID, opts...)
}
