package volume

import (
	"maps"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

// MountMetadataPrefix prefixes metadata keys used for v2 volume mounts.
const MountMetadataPrefix = "volume.v2.sandbox.ucloudai.com:"

// WithMount returns a copy of req mounting the named v2 volume at mountPath.
// It preserves existing metadata and replaces any mount for the same name,
// without modifying the original request or its metadata. Chain calls to
// configure multiple volumes. name is the volume name, not its UUID;
// mountPath is an absolute path inside the sandbox.
//
//	req := volume.WithMount(api.NewSandbox{TemplateID: "base"}, "test-volume", "/mnt")
//	sbx, err := c.Sandboxes().Create(ctx, req, "")
func WithMount(req api.NewSandbox, name, mountPath string) api.NewSandbox {
	metadata := make(api.SandboxMetadata)
	if req.Metadata != nil {
		maps.Copy(metadata, *req.Metadata)
	}
	metadata[MountMetadataPrefix+name] = mountPath
	req.Metadata = &metadata
	return req
}
