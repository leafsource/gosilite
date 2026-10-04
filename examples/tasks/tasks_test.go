package tasks

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/leafsource/gosilite/app"
	"github.com/leafsource/gosilite/dbx"
	"github.com/leafsource/gosilite/httpx"
)

func newTestApp(t *testing.T) (*App, *Service) {
	t.Helper()
	ctx := context.Background()
	db, err := dbx.Open(ctx, "sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sub, err := Migrations()
	if err != nil {
		t.Fatal(err)
	}
	migrations, err := dbx.ParseMigrations(sub)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dbx.Migrate(ctx, db.SQL, db.Dialect, migrations); err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewSQLRepository(db)).WithClock(func() time.Time {
		return time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	})
	return &App{DB: db, Svc: svc}, svc
}

func TestServiceCRUD(t *testing.T) {
	_, svc := newTestApp(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateTaskRequest{Title: "write spec", Description: "docs", Status: "doing"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Title != "write spec" || created.Status != StatusDoing {
		t.Fatalf("created = %+v", created)
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID {
		t.Fatalf("get = %+v", got)
	}

	if _, err := svc.Create(ctx, CreateTaskRequest{Title: ""}); err == nil {
		t.Fatal("expected invalid title")
	} else if e, _ := httpx.AsError(err); e == nil || e.Code != httpx.CodeInvalid {
		t.Fatalf("err = %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, UpdateTaskRequest{Status: strPtr("done")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StatusDone {
		t.Fatalf("updated = %+v", updated)
	}

	list, err := svc.List(ctx, ListTasksQuery{Status: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list = %+v", list)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, created.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestHTTPCRUD(t *testing.T) {
	application, _ := newTestApp(t)
	h := application.HTTPHandler(discardLogger())
	server := httptest.NewServer(h)
	defer server.Close()

	res, body := doJSON(t, server.Client(), http.MethodPost, server.URL+"/api/tasks", `{"title":"api task","status":"todo"}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d body=%s", res.StatusCode, body)
	}
	var created struct {
		OK   bool `json:"ok"`
		Data struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatal(err)
	}
	if !created.OK || created.Data.ID == "" {
		t.Fatalf("created = %s", body)
	}

	res, body = doJSON(t, server.Client(), http.MethodGet, server.URL+"/api/tasks/"+created.Data.ID, "")
	if res.StatusCode != http.StatusOK || !strings.Contains(body, "api task") {
		t.Fatalf("get status=%d body=%s", res.StatusCode, body)
	}

	res, body = doJSON(t, server.Client(), http.MethodPatch, server.URL+"/api/tasks/"+created.Data.ID, `{"status":"done"}`)
	if res.StatusCode != http.StatusOK || !strings.Contains(body, `"done"`) {
		t.Fatalf("patch status=%d body=%s", res.StatusCode, body)
	}

	res, body = doJSON(t, server.Client(), http.MethodGet, server.URL+"/api/tasks?status=done", "")
	if res.StatusCode != http.StatusOK || !strings.Contains(body, "api task") {
		t.Fatalf("list status=%d body=%s", res.StatusCode, body)
	}

	res, body = doJSON(t, server.Client(), http.MethodGet, server.URL+"/api/tasks/nope", "")
	if res.StatusCode != http.StatusNotFound || !strings.Contains(body, `"not_found"`) {
		t.Fatalf("missing status=%d body=%s", res.StatusCode, body)
	}

	res, _ = doJSON(t, server.Client(), http.MethodDelete, server.URL+"/api/tasks/"+created.Data.ID, "")
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", res.StatusCode)
	}

	res, body = doJSON(t, server.Client(), http.MethodPost, server.URL+"/api/tasks", `{"title":"x","nope":true}`)
	if res.StatusCode != http.StatusBadRequest || !strings.Contains(body, `"invalid"`) {
		t.Fatalf("bad body status=%d body=%s", res.StatusCode, body)
	}
}

func TestMCPSameService(t *testing.T) {
	application, _ := newTestApp(t)
	cfg := app.FromEnv()
	cfg.ServiceName = "tasks-test"
	server, err := application.MCPServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tools := server.Tools()
	if len(tools) != 5 {
		t.Fatalf("tools = %d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"tasks_create", "tasks_list", "tasks_get", "tasks_update", "tasks_delete"} {
		if !names[want] {
			t.Fatalf("missing tool %s", want)
		}
	}

	// Call create through the same service the HTTP layer uses.
	created, err := application.Svc.Create(context.Background(), CreateTaskRequest{Title: "via service"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := application.Svc.Get(context.Background(), created.ID)
	if err != nil || got.Title != "via service" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func strPtr(s string) *string { return &s }

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func doJSON(t *testing.T, client *http.Client, method, url, body string) (*http.Response, string) {
	t.Helper()
	var reader io.Reader = strings.NewReader(body)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, string(raw)
}
