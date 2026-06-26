package sandbox

import (
	"context"
	"fmt"
	"net/http"
)

// SetTemplatePublic updates template visibility (private when public is false).
func (c *Client) SetTemplatePublic(ctx context.Context, templateID string, public bool) ([]string, error) {
	if templateID == "" {
		return nil, fmt.Errorf("templateID is required")
	}
	path := fmt.Sprintf("/v2/templates/%s", templateID)
	var resp templateUpdateResponse
	if err := c.doRequest(ctx, http.MethodPatch, path, templateUpdateRequest{Public: public}, &resp); err != nil {
		return nil, err
	}
	return resp.Names, nil
}

// PublishTemplate makes a template public.
func (c *Client) PublishTemplate(ctx context.Context, templateID string) error {
	_, err := c.SetTemplatePublic(ctx, templateID, true)
	return err
}

// UnpublishTemplate makes a template private to the team.
func (c *Client) UnpublishTemplate(ctx context.Context, templateID string) error {
	_, err := c.SetTemplatePublic(ctx, templateID, false)
	return err
}
