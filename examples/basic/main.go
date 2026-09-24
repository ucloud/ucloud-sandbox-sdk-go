// Command basic creates a sandbox, runs a command in it, writes and reads a
// file, then shuts it down.
//
// Set UCLOUD_SANDBOX_API_KEY before running:
//
//	go run ./examples/basic
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

	sbx, err := c.Sandboxes().Create(ctx, api.NewSandbox{
		TemplateID: "system/base",
		Timeout:    new(int32(300)),
		Metadata:   new(api.SandboxMetadata{"example": "basic"}),
	}, "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if _, err := c.Sandboxes().Kill(ctx, sbx.SandboxID); err != nil {
			log.Printf("kill sandbox: %v", err)
		}
	}()

	fmt.Println("sandbox:", sbx.SandboxID)

	envd, err := c.Sandboxes().Envd(sbx, "")
	if err != nil {
		log.Fatal(err)
	}

	out, err := envd.Commands().Run(ctx, "uname -a", commands.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out.Stdout)

	if _, err := envd.Files().Write(ctx, "/home/user/hello.txt", "hello\n"); err != nil {
		log.Fatal(err)
	}

	content, err := envd.Files().Read(ctx, "/home/user/hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("read back: ", content)

	entries, err := envd.Files().List(ctx, "/home/user", 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Printf("%s %s\n", entry.Type, entry.Path)
	}
}
