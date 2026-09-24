// Command secret walks a secret through its whole lifecycle, then shows how a
// sandbox reaches its value without ever holding it.
//
//	go run ./examples/secret
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/sandbox/commands"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/secret"
)

func main() {
	ctx := context.Background()

	c, err := client.New(client.Options{})
	if err != nil {
		log.Fatal(err)
	}

	const name = "example-key"

	created, err := c.Secrets().Create(ctx, api.NewSecret{
		Name:  name,
		Value: "first-value",
		Metadata: new(api.SecretMetadata{
			"example": "secret",
		}),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Secrets().Delete(ctx, created.SecretID)

	fmt.Printf("created %s at version %d\n", created.SecretID, created.CurrentVersion)

	// A secret can be named by ID or by name.
	got, err := c.Secrets().Get(ctx, name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("metadata: %v\n", got.Metadata)

	updated, err := c.Secrets().Update(ctx, name, api.SecretUpdate{
		Value: "second-value",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated to version %d\n", updated.CurrentVersion)

	all, err := c.Secrets().List(ctx, &api.SecretListParams{}).All(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("project has %d secret(s)\n", len(all))

	// The sandbox is given a placeholder, not the value. The runtime resolves
	// it on the way out, so the value is never in the sandbox's environment.
	sbx, err := c.Sandboxes().Create(ctx, api.NewSandbox{
		TemplateID: "base",
		EnvVars:    new(api.EnvVars{"EXAMPLE_KEY": secret.MustFill(name)}),
	}, "")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Sandboxes().Kill(ctx, sbx.SandboxID)

	envd, err := c.Sandboxes().Envd(sbx, "")
	if err != nil {
		log.Fatal(err)
	}

	out, err := envd.Commands().Run(ctx, "printenv EXAMPLE_KEY", commands.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("as seen inside the sandbox: ", out.Stdout)
}
