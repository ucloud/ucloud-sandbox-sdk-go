package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// LogsDirection controls whether logs are returned forward or backward from the cursor.
type LogsDirection string

const (
	LogsDirectionForward  LogsDirection = "forward"
	LogsDirectionBackward LogsDirection = "backward"
)

// LogsSource selects which log store the entries are read from.
type LogsSource string

const (
	LogsSourceTemporary  LogsSource = "temporary"
	LogsSourcePersistent LogsSource = "persistent"
)

// BuildLogEntry is a single structured entry of a template build log.
type BuildLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Step      string    `json:"step,omitempty"`
}

func (e BuildLogEntry) String() string { return e.Message }

type buildLogsConfig struct {
	cursorMs  *int64
	limit     *int
	direction LogsDirection
	level     string
	source    LogsSource
}

// BuildLogsOption customizes a GetTemplateBuildLogs call.
type BuildLogsOption func(*buildLogsConfig)

// WithBuildLogsCursor sets the starting timestamp of the logs, in milliseconds.
func WithBuildLogsCursor(cursorMs int64) BuildLogsOption {
	return func(c *buildLogsConfig) { c.cursorMs = &cursorMs }
}

// WithBuildLogsLimit caps the number of returned entries (1-100, defaults to 100 server-side).
func WithBuildLogsLimit(limit int) BuildLogsOption {
	return func(c *buildLogsConfig) { c.limit = &limit }
}

// WithBuildLogsDirection reads logs forward or backward from the cursor.
func WithBuildLogsDirection(direction LogsDirection) BuildLogsOption {
	return func(c *buildLogsConfig) { c.direction = direction }
}

// WithBuildLogsLevel filters entries by minimum severity (debug, info, warn, error).
func WithBuildLogsLevel(level string) BuildLogsOption {
	return func(c *buildLogsConfig) { c.level = level }
}

// WithBuildLogsSource selects the temporary or persistent log store.
func WithBuildLogsSource(source LogsSource) BuildLogsOption {
	return func(c *buildLogsConfig) { c.source = source }
}

type templateBuildLogsResponse struct {
	Logs []BuildLogEntry `json:"logs"`
}

// GetTemplateBuildLogs returns the structured build logs of a template build.
func (c *Client) GetTemplateBuildLogs(ctx context.Context, templateID, buildID string, opts ...BuildLogsOption) ([]BuildLogEntry, error) {
	if templateID == "" {
		return nil, fmt.Errorf("templateID is required")
	}
	if buildID == "" {
		return nil, fmt.Errorf("buildID is required")
	}

	cfg := &buildLogsConfig{}
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
	if cfg.source != "" {
		query.Set("source", string(cfg.source))
	}

	path := fmt.Sprintf("/templates/%s/builds/%s/logs", url.PathEscape(templateID), url.PathEscape(buildID))
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp templateBuildLogsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("get template build logs %s/%s: %w", templateID, buildID, err)
	}
	return resp.Logs, nil
}
