package mcpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/leafsource/gosilite/httpx"
)

func TestServerToolsListAndCall(t *testing.T) {
	s := NewServer("test", "1.0.0", nil)
	err := s.Register(Tool{
		Name:        "echo",
		Description: "echo args",
		InputSchema: ObjectSchema(map[string]any{"text": StringProp("text")}, "text"),
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, httpx.Invalid("bad args")
			}
			return map[string]string{"echo": in.Text}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Register(Tool{Name: "echo", Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
		return nil, nil
	}}); err == nil {
		t.Fatal("expected duplicate name error")
	}

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"text":"hi"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"text":"bad"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"missing","arguments":{}}}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	if err := s.ServeStdio(context.Background(), strings.NewReader(in), &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines:\n%s", len(lines), out.String())
	}

	var list struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Result.Tools) != 1 || list.Result.Tools[0].Name != "echo" {
		t.Fatalf("tools/list = %+v", list)
	}

	var callOK struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &callOK); err != nil {
		t.Fatal(err)
	}
	if callOK.Result.IsError {
		t.Fatalf("unexpected error: %s", lines[2])
	}
	if !strings.Contains(callOK.Result.Content[0].Text, `"hi"`) {
		t.Fatalf("payload = %s", callOK.Result.Content[0].Text)
	}
}

func TestToolCallBusinessError(t *testing.T) {
	s := NewServer("test", "1.0.0", nil)
	_ = s.Register(Tool{
		Name: "boom",
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			return nil, httpx.NotFound("item not found")
		},
	})
	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"boom","arguments":{}}}` + "\n"
	var out bytes.Buffer
	if err := s.ServeStdio(context.Background(), strings.NewReader(in), &out); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Result.IsError {
		t.Fatal("expected isError")
	}
	if !strings.Contains(resp.Result.Content[0].Text, `"not_found"`) {
		t.Fatalf("payload = %s", resp.Result.Content[0].Text)
	}
}

func TestErrorPayloadInternal(t *testing.T) {
	m := errorPayload(errors.New("db down"))
	if m["code"] != httpx.CodeInternal {
		t.Fatalf("code = %v", m["code"])
	}
}
