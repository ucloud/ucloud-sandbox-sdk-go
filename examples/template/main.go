// Command template builds a template and starts a sandbox from it.
//
//	go run ./examples/template
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/template/build"
)

func main() {
	ctx := context.Background()

	c, err := client.New(client.Options{})
	if err != nil {
		log.Fatal(err)
	}

	builder := build.FromBaseImage().
		RunCmd("apt-get update && apt-get install -y python3").
		SetWorkdir("/app").
		SetLogger(build.DefaultLogger())

	build, err := builder.Build(ctx, c, api.TemplateBuildRequestV3{
		Name:     new("exmaple-python"),
		CpuCount: new(int32(2)),
		MemoryMB: new(int32(2048)),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("built template:", build.TemplateID)

	sbx, err := c.Sandboxes().Create(ctx, api.NewSandbox{
		TemplateID: build.TemplateID,
	}, "")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Sandboxes().Kill(ctx, sbx.SandboxID)

	envd, err := c.Sandboxes().Envd(sbx, "")
	if err != nil {
		log.Fatal(err)
	}

	out, err := envd.Commands().Run(ctx, "python3 --version", commands.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out.Stdout)
}
