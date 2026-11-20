// Command volume-v2 creates a cloud NFS volume and mounts it into a sandbox.
//
//	go run ./examples/volume-v2
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
	volume "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2"
	volumeapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	c, err := client.New(client.Options{})
	if err != nil {
		return err
	}
	v, err := c.VolumesV2().Create(ctx, volumeapi.CreateVolume{
		Name: "example-data-v2", SizeGiB: 100,
		ExtraParams: new(volumeapi.ExtraParams{"storageType": "Advanced"}),
	}, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.VolumesV2().Delete(ctx, v.VolumeID); err != nil {
			log.Printf("delete volume: %v", err)
		}
	}()
	fmt.Println("volume:", v.VolumeID)

	req := volume.WithMount(api.NewSandbox{TemplateID: "base"}, v.Name, "/mnt")
	sbx, err := c.Sandboxes().Create(ctx, req, "")
	if err != nil {
		return err
	}
	defer c.Sandboxes().Kill(ctx, sbx.SandboxID)

	out, err := sbx.Commands().Run(ctx,
		"echo persisted > /mnt/note.txt && cat /mnt/note.txt", commands.Options{})
	if err != nil {
		return err
	}
	fmt.Print(out.Stdout)
	return nil
}
