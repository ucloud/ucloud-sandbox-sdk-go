package sandbox

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSandboxLogs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/sandboxes/sbx-1/logs" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"cursor":    "1754000000000",
			"limit":     "500",
			"direction": "forward",
			"level":     "warn",
			"search":    "OOM killed",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"logs":[
			{"timestamp":"2026-08-01T10:00:00Z","message":"OOM killed","level":"error","fields":{"service":"envd","pid":"42"}}
		]}`)
	}))
	defer server.Close()

	client := NewClient("example.com", "api-key", WithAPIURL(server.URL))

	logs, err := client.GetSandboxLogs(context.Background(), "sbx-1",
		WithSandboxLogsCursor(1754000000000),
		WithSandboxLogsLimit(500),
		WithSandboxLogsDirection(LogsDirectionForward),
		WithSandboxLogsLevel("warn"),
		WithSandboxLogsSearch("OOM killed"),
	)
	if err != nil {
		t.Fatalf("GetSandboxLogs: %v", err)
	}
	if got, want := len(logs), 1; got != want {
		t.Fatalf("len(logs) = %d, want %d", got, want)
	}
	if got, want := logs[0].Message, "OOM killed"; got != want {
		t.Errorf("Message = %q, want %q", got, want)
	}
	if got, want := logs[0].Level, "error"; got != want {
		t.Errorf("Level = %q, want %q", got, want)
	}
	if got, want := logs[0].Fields["service"], "envd"; got != want {
		t.Errorf("Fields[service] = %q, want %q", got, want)
	}
	if got, want := logs[0].String(), "OOM killed"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestGetSandboxLogsNoOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("RawQuery = %q, want empty", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"logs":[]}`)
	}))
	defer server.Close()

	client := NewClient("example.com", "api-key", WithAPIURL(server.URL))

	logs, err := client.GetSandboxLogs(context.Background(), "sbx-1")
	if err != nil {
		t.Fatalf("GetSandboxLogs: %v", err)
	}
	if logs == nil {
		t.Error("logs = nil, want empty slice")
	}
	if got, want := len(logs), 0; got != want {
		t.Errorf("len(logs) = %d, want %d", got, want)
	}
}

func TestGetSandboxLogsRequiresID(t *testing.T) {
	t.Parallel()

	client := NewClient("example.com", "api-key", WithAPIURL("http://127.0.0.1:1"))
	if _, err := client.GetSandboxLogs(context.Background(), ""); err == nil {
		t.Error("GetSandboxLogs(\"\") = nil error, want error")
	}
}
