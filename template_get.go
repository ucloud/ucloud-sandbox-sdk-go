package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// nextTokenHeader carries the pagination cursor of GetTemplate responses.
const nextTokenHeader = "X-Next-Token"

// TemplateBuild describes a single build of a template.
type TemplateBuild struct {
	BuildID     string              `json:"buildID"`
	Status      TemplateBuildStatus `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	FinishedAt  *time.Time          `json:"finishedAt,omitempty"`
	CPUCount    int                 `json:"cpuCount"`
	MemoryMB    int                 `json:"memoryMB"`
	DiskSizeMB  int                 `json:"diskSizeMB,omitempty"`
	EnvdVersion string              `json:"envdVersion,omitempty"`
}

// TemplateWithBuilds is a template together with a page of its builds.
type TemplateWithBuilds struct {
	TemplateID    string          `json:"templateID"`
	Public        bool            `json:"public"`
	Aliases       []string        `json:"aliases,omitempty"`
	Names         []string        `json:"names,omitempty"` // namespace/alias format when namespaced
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	LastSpawnedAt *time.Time      `json:"lastSpawnedAt,omitempty"`
	SpawnCount    int64           `json:"spawnCount"`
	Builds        []TemplateBuild `json:"builds"`

	// NextToken is the cursor for the next page of builds, empty when the last
	// page has been reached. It comes from the X-Next-Token response header.
	NextToken string `json:"-"`
}

type getTemplateConfig struct {
	nextToken string
	limit     *int
}

// GetTemplateOption customizes a GetTemplate call.
type GetTemplateOption func(*getTemplateConfig)

// WithTemplateBuildsNextToken starts the builds list from the given cursor.
func WithTemplateBuildsNextToken(token string) GetTemplateOption {
	return func(c *getTemplateConfig) { c.nextToken = token }
}

// WithTemplateBuildsLimit caps the number of builds per page (1-100, defaults to 100 server-side).
func WithTemplateBuildsLimit(limit int) GetTemplateOption {
	return func(c *getTemplateConfig) { c.limit = &limit }
}

// GetTemplate returns a template along with a page of its builds.
func (c *Client) GetTemplate(ctx context.Context, templateID string, opts ...GetTemplateOption) (*TemplateWithBuilds, error) {
	if templateID == "" {
		return nil, fmt.Errorf("templateID is required")
	}

	cfg := &getTemplateConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	query := url.Values{}
	if cfg.nextToken != "" {
		query.Set("nextToken", cfg.nextToken)
	}
	if cfg.limit != nil {
		query.Set("limit", strconv.Itoa(*cfg.limit))
	}

	path := "/templates/" + url.PathEscape(templateID)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var result TemplateWithBuilds
	headers, err := c.doRequestWithHeaders(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("get template %s: %w", templateID, err)
	}
	if headers != nil {
		result.NextToken = headers.Get(nextTokenHeader)
	}
	return &result, nil
}

// ListTemplateBuilds paginates over all builds of a template.
func (c *Client) ListTemplateBuilds(ctx context.Context, templateID string, opts ...GetTemplateOption) *Paginator[TemplateBuild] {
	cfg := &getTemplateConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	limit := 0
	if cfg.limit != nil {
		limit = *cfg.limit
	}

	return newPaginator(limit, func(ctx context.Context, token string, limit int) ([]TemplateBuild, string, error) {
		pageOpts := make([]GetTemplateOption, 0, 2)
		if token == "" {
			token = cfg.nextToken
		}
		if token != "" {
			pageOpts = append(pageOpts, WithTemplateBuildsNextToken(token))
		}
		if limit > 0 {
			pageOpts = append(pageOpts, WithTemplateBuildsLimit(limit))
		}
		tpl, err := c.GetTemplate(ctx, templateID, pageOpts...)
		if err != nil {
			return nil, "", err
		}
		return tpl.Builds, tpl.NextToken, nil
	})
}
