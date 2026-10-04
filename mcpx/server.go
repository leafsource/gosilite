package mcpx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/leafsource/gosilite/httpx"
)

const protocolVersion = "2024-11-05"

// Server is a minimal MCP tools server over JSON-RPC 2.0 (stdio).
// It exposes business capabilities only — not system/ops tools.
type Server struct {
	Name    string
	Version string
	tools   []Tool
	byName  map[string]Tool
	logger  *slog.Logger
}

// NewServer creates an empty business MCP server.
func NewServer(name, version string, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		Name:    name,
		Version: version,
		byName:  map[string]Tool{},
		logger:  logger,
	}
}

// Register adds a business tool. Duplicate names are rejected.
func (s *Server) Register(t Tool) error {
	if t.Name == "" {
		return fmt.Errorf("mcpx: tool name is required")
	}
	if t.Handler == nil {
		return fmt.Errorf("mcpx: tool %q handler is required", t.Name)
	}
	if _, exists := s.byName[t.Name]; exists {
		return fmt.Errorf("mcpx: tool %q already registered", t.Name)
	}
	if t.InputSchema == nil {
		t.InputSchema = ObjectSchema(map[string]any{})
	}
	s.tools = append(s.tools, t)
	s.byName[t.Name] = t
	return nil
}

// Tools returns registered tools in registration order.
func (s *Server) Tools() []Tool {
	out := make([]Tool, len(s.tools))
	copy(out, s.tools)
	return out
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type toolCallResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ServeStdio processes newline-delimited JSON-RPC messages until EOF or ctx cancel.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			if handleErr := s.handleLine(ctx, line, writer); handleErr != nil {
				return handleErr
			}
			_ = writer.Flush()
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (s *Server) handleLine(ctx context.Context, line []byte, out *bufio.Writer) error {
	var req rpcRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &rpcError{Code: -32700, Message: "parse error"},
		})
	}
	// Notifications (no id) are accepted and ignored except initialize is a request.
	if req.Method == "" {
		return nil
	}

	switch req.Method {
	case "initialize":
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": protocolVersion,
				"capabilities": map[string]any{
					"tools": map[string]any{"listChanged": false},
				},
				"serverInfo": map[string]any{
					"name":    s.Name,
					"version": s.Version,
				},
			},
		})
	case "notifications/initialized", "notifications/cancelled":
		return nil
	case "ping":
		return s.write(out, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}})
	case "tools/list":
		list := make([]toolDescriptor, 0, len(s.tools))
		for _, t := range s.tools {
			list = append(list, toolDescriptor{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"tools": list},
		})
	case "tools/call":
		return s.handleCall(ctx, req, out)
	default:
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
		})
	}
}

func (s *Server) handleCall(ctx context.Context, req rpcRequest, out *bufio.Writer) error {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "invalid params"},
		})
	}
	tool, ok := s.byName[params.Name]
	if !ok {
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "unknown tool: " + params.Name},
		})
	}
	args := params.Arguments
	if len(args) == 0 {
		args = json.RawMessage(`{}`)
	}
	result, err := tool.Handler(ctx, args)
	if err != nil {
		payload := map[string]any{
			"ok":    false,
			"error": errorPayload(err),
		}
		raw, _ := json.Marshal(payload)
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: toolCallResult{
				IsError: true,
				Content: []toolContent{{Type: "text", Text: string(raw)}},
			},
		})
	}
	raw, err := json.Marshal(map[string]any{"ok": true, "data": result})
	if err != nil {
		return s.write(out, rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32603, Message: "marshal result failed"},
		})
	}
	return s.write(out, rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: toolCallResult{
			Content: []toolContent{{Type: "text", Text: string(raw)}},
		},
	})
}

func errorPayload(err error) map[string]any {
	// Keep code parity with HTTP so agent clients share one error vocabulary.
	if e, ok := httpx.AsError(err); ok {
		return map[string]any{
			"code":    e.Code,
			"message": e.Message,
			"detail":  e.Detail,
		}
	}
	return map[string]any{
		"code":    httpx.CodeInternal,
		"message": err.Error(),
	}
}

func (s *Server) write(out *bufio.Writer, resp rpcResponse) error {
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	_, err = out.Write(raw)
	return err
}
