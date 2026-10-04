package mcpx

import (
	"context"
	"encoding/json"
)

// Tool is one business capability exposed over MCP.
// It wraps the same Service used by HTTP handlers.
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	Handler     ToolHandler
}

// ToolHandler executes a business call. args is the JSON object from the MCP client.
type ToolHandler func(ctx context.Context, args json.RawMessage) (any, error)

// objectSchema builds a simple JSON Schema for an object with required fields.
func ObjectSchema(props map[string]any, required ...string) json.RawMessage {
	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	raw, _ := json.Marshal(schema)
	return raw
}

// StringProp is a convenience JSON Schema string property.
func StringProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

// StringEnumProp is a JSON Schema string enum property.
func StringEnumProp(desc string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": desc, "enum": values}
}
