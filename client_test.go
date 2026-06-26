package sandbox

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func testCredentials(t *testing.T) (domain, apiKey string) {
	t.Helper()
	domain = os.Getenv("UCLOUD_SANDBOX_DOMAIN")
	apiKey = os.Getenv("UCLOUD_SANDBOX_API_KEY")
	if domain == "" || apiKey == "" {
		t.Skip("set UCLOUD_SANDBOX_DOMAIN and UCLOUD_SANDBOX_API_KEY to run integration test")
	}
	return domain, apiKey
}

func TestSandboxLifecycle(t *testing.T) {
	fmtTime := func(ts time.Time) string {
		return ts.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")
	}
	domain, apiKey := testCredentials(t)
	ctx := context.Background()
	client := NewClient(domain, apiKey)

	existing, _ := client.ListSandboxes(ctx, nil).All(ctx)
	fmt.Printf("existing sandboxes count=%d\n", len(existing))

	sbx, err := client.CreateSandbox(ctx,
		WithTemplate("base"),
		WithTimeout(5),
		WithAutoPause(true),
		WithAutoResume(AutoResumePolicyOn),
	)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() {
		client.KillSandbox(ctx, sbx.ID)
	}()

	result, err := sbx.Commands.Run(ctx, "echo hello", WithCommandTimeout(30))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("unexpected exit code %d", result.ExitCode)
	}

	info, err := client.GetSandboxInfo(ctx, sbx.ID)
	if err != nil {
		t.Fatalf("get info: %v", err)
	}
	fmt.Printf("sandbox id=%s state=%s endAt=%s stdout=%s\n", info.SandboxID, info.State, fmtTime(info.EndAt), result.Stdout)
}
