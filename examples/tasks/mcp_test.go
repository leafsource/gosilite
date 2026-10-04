package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leafsource/gosilite/app"
)

func TestMCPBusinessToolsEndToEnd(t *testing.T) {
	application, _ := newTestApp(t)
	handler := application.HTTPHandler(slog.New(slog.NewTextHandler(new(bytes.Buffer), nil)))
	server := httptest.NewServer(handler)
	defer server.Close()
	cfg := app.FromEnv()
	cfg.ServiceName = "tasks-mcp"
	mcpServer, err := application.MCPServer(cfg)
	if err != nil {
		t.Fatal(err)
	}

	script := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"tasks_create","arguments":{"title":"mcp task","status":"todo"}}}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	if err := mcpServer.ServeStdio(context.Background(), strings.NewReader(script), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines=%d out=%s", len(lines), out.String())
	}

	var listResp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Result.Tools) != 5 {
		t.Fatalf("tools=%d", len(listResp.Result.Tools))
	}

	var callResp struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &callResp); err != nil {
		t.Fatal(err)
	}
	if callResp.Result.IsError {
		t.Fatalf("create failed: %s", lines[2])
	}
	var created struct {
		Data struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(callResp.Result.Content[0].Text), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == "" || created.Data.Title != "mcp task" {
		t.Fatalf("payload=%s", callResp.Result.Content[0].Text)
	}

	// HTTP can read what MCP created (shared service/store).
	res, body := doJSON(t, server.Client(), http.MethodGet, server.URL+"/api/tasks/"+created.Data.ID, "")
	if res.StatusCode != http.StatusOK || !strings.Contains(body, "mcp task") {
		t.Fatalf("http get status=%d body=%s", res.StatusCode, body)
	}

	// tools/call get/update/delete against the same ids.
	script2 := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"tasks_get","arguments":{"id":"` + created.Data.ID + `"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"tasks_update","arguments":{"id":"` + created.Data.ID + `","status":"done"}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"tasks_delete","arguments":{"id":"` + created.Data.ID + `"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"tasks_get","arguments":{"id":"` + created.Data.ID + `"}}}`,
	}, "\n") + "\n"
	out.Reset()
	if err := mcpServer.ServeStdio(context.Background(), strings.NewReader(script2), &out); err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("lines=%d out=%s", len(lines), out.String())
	}
	if !strings.Contains(lines[0], "mcp task") {
		t.Fatalf("get = %s", lines[0])
	}
	var updateResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &updateResp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updateResp.Result.Content[0].Text, `"status":"done"`) {
		t.Fatalf("update = %s", lines[1])
	}
	var deleteResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &deleteResp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(deleteResp.Result.Content[0].Text, `"deleted"`) {
		t.Fatalf("delete = %s", lines[2])
	}
	if !strings.Contains(lines[3], "not_found") {
		t.Fatalf("get-after-delete = %s", lines[3])
	}

	// Shared store empty after MCP delete.
	tasks, err := application.Svc.List(context.Background(), ListTasksQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Fatalf("shared store = %+v", tasks)
	}
}

func TestMCPServerRegistersBusinessOnly(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(new(bytes.Buffer), nil))
	_ = logger
	application, _ := newTestApp(t)
	server, err := application.MCPServer(app.FromEnv())
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range server.Tools() {
		if !strings.HasPrefix(tool.Name, "tasks_") {
			t.Fatalf("unexpected non-business tool %q", tool.Name)
		}
	}
}
