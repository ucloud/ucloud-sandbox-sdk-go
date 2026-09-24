// Command volume creates a volume and mounts it into a sandbox.
//
//	go run ./examples/volume
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

	c, err := client.New(client.Options{})
	if err != nil {
		log.Fatal(err)
	}

	vol, err := c.Volumes().Create(ctx, "example-data")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Volumes().Delete(ctx, vol.VolumeID)

	fmt.Println("volume:", vol.VolumeID)

	sbx, err := c.Sandboxes().Create(ctx, api.NewSandbox{
		TemplateID: "base",
		VolumeMounts: new([]api.SandboxVolumeMount{
			{
				Name: vol.Name,
				Path: "/mnt/data",
			},
		}),
	}, "")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Sandboxes().Kill(ctx, sbx.SandboxID)

	envd, err := c.Sandboxes().Envd(sbx, "")
	if err != nil {
		log.Fatal(err)
	}

	// The volume is written from inside the sandbox, where it is mounted.
	out, err := envd.Commands().Run(ctx,
		"echo persisted > /mnt/data/note.txt && cat /mnt/data/note.txt",
		commands.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out.Stdout)
}
