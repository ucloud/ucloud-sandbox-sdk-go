package volume_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox"
	volume "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2"
)

func TestWithMount(t *testing.T) {
	for _, metadata := range []*api.SandboxMetadata{nil, new(api.SandboxMetadata(nil)), new(api.SandboxMetadata{"app": "test"})} {
		original := api.NewSandbox{TemplateID: "base", Metadata: metadata}
		mounted := volume.WithMount(original, "test-volume", "/mnt")
		require.NotNil(t, mounted.Metadata)
		assert.Equal(t, "/mnt", (*mounted.Metadata)["volume.v2.sandbox.ucloudai.com:test-volume"])
		assert.Equal(t, "base", mounted.TemplateID)
		if original.Metadata != nil {
			assert.NotContains(t, *original.Metadata, "volume.v2.sandbox.ucloudai.com:test-volume")
			if app, ok := (*original.Metadata)["app"]; ok {
				assert.Equal(t, app, (*mounted.Metadata)["app"])
			}
		}
		second := volume.WithMount(mounted, "second-volume", "/data")
		replaced := volume.WithMount(second, "test-volume", "/new-mnt")
		assert.Equal(t, "/new-mnt", (*replaced.Metadata)["volume.v2.sandbox.ucloudai.com:test-volume"])
		assert.Equal(t, "/data", (*replaced.Metadata)["volume.v2.sandbox.ucloudai.com:second-volume"])
		assert.Equal(t, "/mnt", (*mounted.Metadata)["volume.v2.sandbox.ucloudai.com:test-volume"])
		assert.NotContains(t, *mounted.Metadata, "volume.v2.sandbox.ucloudai.com:second-volume")
	}
}

func TestMountSentOnSandboxCreation(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/gateway/sandboxes", r.URL.Path)
		var body api.NewSandbox
		if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) && assert.NotNil(t, body.Metadata) {
			assert.Equal(t, api.SandboxMetadata{
				"app": "test",
				"volume.v2.sandbox.ucloudai.com:test-volume": "/mnt",
				sandbox.ManageByMetadataKey:                  sandbox.ManageByDefault,
			}, *body.Metadata)
			assert.Nil(t, body.VolumeMounts)
		}
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 400, "message": "test response"}))
	})
	req := api.NewSandbox{TemplateID: "base", Metadata: new(api.SandboxMetadata{"app": "test"})}
	mounted := volume.WithMount(req, "test-volume", "/mnt")
	_, err := c.Sandboxes().Create(context.Background(), mounted, "")
	require.ErrorIs(t, err, errdefs.ErrInvalidArgument)
	assert.Equal(t, api.SandboxMetadata{"app": "test"}, *req.Metadata)
	assert.NotContains(t, *mounted.Metadata, sandbox.ManageByMetadataKey)
}
