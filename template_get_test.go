package sandbox

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTemplateWithBuilds(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/templates/tpl-1" {
			http.NotFound(w, r)
			return
		}
		if got, want := r.URL.Query().Get("limit"), "2"; got != want {
			t.Errorf("limit = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Next-Token", "cursor-2")
		_, _ = io.WriteString(w, `{
			"templateID": "tpl-1",
			"public": false,
			"names": ["team/base"],
			"createdAt": "2026-08-01T10:00:00Z",
			"updatedAt": "2026-08-02T10:00:00Z",
			"lastSpawnedAt": null,
			"spawnCount": 7,
			"builds": [
				{"buildID":"b-1","status":"ready","createdAt":"2026-08-01T10:00:00Z","updatedAt":"2026-08-01T10:05:00Z","finishedAt":"2026-08-01T10:05:00Z","cpuCount":2,"memoryMB":1024,"diskSizeMB":2048,"envdVersion":"0.1.5"},
				{"buildID":"b-2","status":"error","createdAt":"2026-08-02T10:00:00Z","updatedAt":"2026-08-02T10:01:00Z","cpuCount":4,"memoryMB":2048}
			]
		}`)
	}))
	defer server.Close()

	client := NewClient("example.com", "api-key", WithAPIURL(server.URL))

	tpl, err := client.GetTemplate(context.Background(), "tpl-1", WithTemplateBuildsLimit(2))
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if got, want := tpl.TemplateID, "tpl-1"; got != want {
		t.Errorf("TemplateID = %q, want %q", got, want)
	}
	if got, want := tpl.SpawnCount, int64(7); got != want {
		t.Errorf("SpawnCount = %d, want %d", got, want)
	}
	if tpl.LastSpawnedAt != nil {
		t.Errorf("LastSpawnedAt = %v, want nil", tpl.LastSpawnedAt)
	}
	if got, want := tpl.NextToken, "cursor-2"; got != want {
		t.Errorf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(tpl.Builds), 2; got != want {
		t.Fatalf("len(Builds) = %d, want %d", got, want)
	}
	if got, want := tpl.Builds[0].Status, BuildStatusReady; got != want {
		t.Errorf("Builds[0].Status = %q, want %q", got, want)
	}
	if tpl.Builds[0].FinishedAt == nil {
		t.Error("Builds[0].FinishedAt = nil, want a timestamp")
	}
	if got, want := tpl.Builds[1].MemoryMB, 2048; got != want {
		t.Errorf("Builds[1].MemoryMB = %d, want %d", got, want)
	}
}

func TestListTemplateBuildsPaginates(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/tpl-1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("nextToken") {
		case "":
			w.Header().Set("X-Next-Token", "page-2")
			_, _ = io.WriteString(w, `{"templateID":"tpl-1","builds":[{"buildID":"b-1","status":"ready","cpuCount":2,"memoryMB":1024}]}`)
		case "page-2":
			_, _ = io.WriteString(w, `{"templateID":"tpl-1","builds":[{"buildID":"b-2","status":"ready","cpuCount":2,"memoryMB":1024}]}`)
		default:
			t.Errorf("unexpected nextToken %q", r.URL.Query().Get("nextToken"))
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient("example.com", "api-key", WithAPIURL(server.URL))

	builds, err := client.ListTemplateBuilds(context.Background(), "tpl-1").All(context.Background())
	if err != nil {
		t.Fatalf("ListTemplateBuilds: %v", err)
	}
	if got, want := len(builds), 2; got != want {
		t.Fatalf("len(builds) = %d, want %d", got, want)
	}
	if got, want := builds[1].BuildID, "b-2"; got != want {
		t.Errorf("builds[1].BuildID = %q, want %q", got, want)
	}
}

func TestGetTemplateBuildLogs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/templates/tpl-1/builds/b-1/logs" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"cursor":    "1754000000000",
			"limit":     "50",
			"direction": "backward",
			"level":     "error",
			"source":    "persistent",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"logs":[
			{"timestamp":"2026-08-01T10:00:00Z","message":"step failed","level":"error","step":"RUN apt-get update"}
		]}`)
	}))
	defer server.Close()

	client := NewClient("example.com", "api-key", WithAPIURL(server.URL))

	logs, err := client.GetTemplateBuildLogs(context.Background(), "tpl-1", "b-1",
		WithBuildLogsCursor(1754000000000),
		WithBuildLogsLimit(50),
		WithBuildLogsDirection(LogsDirectionBackward),
		WithBuildLogsLevel("error"),
		WithBuildLogsSource(LogsSourcePersistent),
	)
	if err != nil {
		t.Fatalf("GetTemplateBuildLogs: %v", err)
	}
	if got, want := len(logs), 1; got != want {
		t.Fatalf("len(logs) = %d, want %d", got, want)
	}
	if got, want := logs[0].Message, "step failed"; got != want {
		t.Errorf("Message = %q, want %q", got, want)
	}
	if got, want := logs[0].Step, "RUN apt-get update"; got != want {
		t.Errorf("Step = %q, want %q", got, want)
	}
	if got, want := logs[0].String(), "step failed"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestGetTemplateRequiresID(t *testing.T) {
	t.Parallel()

	client := NewClient("example.com", "api-key", WithAPIURL("http://127.0.0.1:1"))
	if _, err := client.GetTemplate(context.Background(), ""); err == nil {
		t.Error("GetTemplate(\"\") = nil error, want error")
	}
	if _, err := client.GetTemplateBuildLogs(context.Background(), "tpl-1", ""); err == nil {
		t.Error("GetTemplateBuildLogs with empty buildID = nil error, want error")
	}
}
