# ucloud-sandbox-sdk-go

Go SDK for UCloud Sandbox: cloud VMs that boot in about a second, for running
code you did not write.

## Install

```bash
go get github.com/ucloud/ucloud-sandbox-sdk-go
```

Requires Go 1.27 or newer.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
)

func main() {
	ctx := context.Background()

	c, err := client.New(client.Options{APIKey: "your-api-key"})
	if err != nil {
		log.Fatal(err)
	}

	sbx, err := c.Sandboxes().Create(ctx, api.NewSandbox{TemplateID: "system/base"})
	if err != nil {
		log.Fatal(err)
	}
	defer sbx.Kill(ctx)

	out, err := sbx.Commands().Run(ctx, "uname -a", commands.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out.Stdout)
}
```

Runnable versions of this and more are in [`examples/`](examples/).

## Configuration

`client.New` takes everything through `client.Options`. Leave a field empty and
it falls back to an environment variable, then to a default:

| Field | Environment variable | Default |
| --- | --- | --- |
| `APIKey` | `UCLOUD_SANDBOX_API_KEY` | required — `New` fails without it |
| `Region` | `UCLOUD_SANDBOX_REGION` | `cn-wlcb` |
| `Domain` | `UCLOUD_SANDBOX_DOMAIN` | `{region}.sandbox.ucloudai.com` |
| `APIURL` | `UCLOUD_SANDBOX_API_URL` | `https://api.{domain}` |
| `VolumeContentURL` | `UCLOUD_SANDBOX_VOLUME_CONTENT_URL` | `https://volumecontent.{domain}` |
| `InsecureHTTP` | `UCLOUD_SANDBOX_INSECURE_HTTP` | `false` |

`Region` expands to a domain; `Domain` overrides it; `APIURL` overrides both, so
a private deployment can be named by address. `VolumeContentURL` can likewise
point to a hostname or IP address:

```go
c, err := client.New(client.Options{
	APIKey:           apiKey,
	APIURL:           "http://10.10.0.5:8080",
	VolumeContentURL: "http://10.10.0.5:8081",
	InsecureHTTP:     true,
})
```

Set `UCLOUD_SANDBOX_VOLUME_CONTENT_URL` to configure the volume content endpoint
through the environment instead.

## Packages

| Package | What it does |
| --- | --- |
| [`pkg/client`](pkg/client) | Entry point. `New` plus the services below. |
| [`pkg/sandbox`](pkg/sandbox) | Sandboxes, and the commands, files and PTYs inside them. |
| [`pkg/template`](pkg/template) | Building the images sandboxes boot from. |
| [`pkg/volume`](pkg/volume) | Volume management, including capacity and usage. |
| [`pkg/volume/content`](pkg/volume/content) | Reading and writing files in a volume. |
| [`pkg/secret`](pkg/secret) | Project secrets. |
| [`pkg/errdefs`](pkg/errdefs) | Error types, and the mappings that produce them. |
| [`pkg/transport`](pkg/transport) | HTTP plumbing. Rarely used directly. |
| [`pkg/api`](pkg/api), [`pkg/envd`](pkg/envd) | Generated clients. See below. |

### Errors

Every failure, whether it came over HTTP or over envd's RPC, lands on the same
types:

```go
if errors.Is(err, errdefs.ErrNotFound) { ... }

var exit *errdefs.CommandExitError
if errors.As(err, &exit) {
	log.Print(exit.Stderr)
}
```

## Development

The clients in `pkg/api`, `pkg/envd`, and `pkg/volume` are generated. The
volume clients can be regenerated with `make generate-volume`; regenerating all
clients needs the submodules:

```bash
git submodule update --init
make generate
make test
```

The generated code is committed, so `go get` works without any of this.

`spec/openapi.yml` describes the control plane. Volume management and content
clients are generated from `spec/openapi-volume.yaml` and
`spec/openapi-volumecontent.yaml`; run `make generate-volume` to regenerate
only those clients. envd's specs come from
[e2b-dev/runtime](https://github.com/e2b-dev/runtime), vendored under
`submodules/` and pinned to a tag; see [`pkg/envd/README.md`](pkg/envd/README.md).
`submodules/ucloud-sandbox-sdk-python` is the Python SDK, kept only as a
reference for behaviour the two should share.

## Licence

MIT. `pkg/envd` is generated from Apache-2.0 specs; see
[`pkg/envd/LICENSE-e2b`](pkg/envd/LICENSE-e2b).
