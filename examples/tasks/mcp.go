package tasks

import (
	"context"
	"encoding/json"

	"github.com/leafsource/gosilite/httpx"
	"github.com/leafsource/gosilite/mcpx"
)

// MCPTools registers business capabilities (same Service as HTTP) as MCP tools.
func MCPTools(svc *Service) ([]mcpx.Tool, error) {
	tools := []mcpx.Tool{
		{
			Name:        "tasks_create",
			Description: "Create a task with title, optional description and status (todo|doing|done).",
			InputSchema: mcpx.ObjectSchema(map[string]any{
				"title":       mcpx.StringProp("Task title"),
				"description": mcpx.StringProp("Optional description"),
				"status":      mcpx.StringEnumProp("Initial status", "todo", "doing", "done"),
			}, "title"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var in CreateTaskRequest
				if err := json.Unmarshal(args, &in); err != nil {
					return nil, httpx.Invalid("invalid arguments")
				}
				return svc.Create(ctx, in)
			},
		},
		{
			Name:        "tasks_list",
			Description: "List tasks, optionally filter by status.",
			InputSchema: mcpx.ObjectSchema(map[string]any{
				"status": mcpx.StringEnumProp("Filter status", "todo", "doing", "done"),
			}),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var in ListTasksQuery
				if len(args) > 0 {
					if err := json.Unmarshal(args, &in); err != nil {
						return nil, httpx.Invalid("invalid arguments")
					}
				}
				return svc.List(ctx, in)
			},
		},
		{
			Name:        "tasks_get",
			Description: "Get one task by id.",
			InputSchema: mcpx.ObjectSchema(map[string]any{
				"id": mcpx.StringProp("Task id"),
			}, "id"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var in struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(args, &in); err != nil {
					return nil, httpx.Invalid("invalid arguments")
				}
				return svc.Get(ctx, in.ID)
			},
		},
		{
			Name:        "tasks_update",
			Description: "Partially update a task (title/description/status).",
			InputSchema: mcpx.ObjectSchema(map[string]any{
				"id":          mcpx.StringProp("Task id"),
				"title":       mcpx.StringProp("New title"),
				"description": mcpx.StringProp("New description"),
				"status":      mcpx.StringEnumProp("New status", "todo", "doing", "done"),
			}, "id"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var in struct {
					ID string `json:"id"`
					UpdateTaskRequest
				}
				if err := json.Unmarshal(args, &in); err != nil {
					return nil, httpx.Invalid("invalid arguments")
				}
				return svc.Update(ctx, in.ID, in.UpdateTaskRequest)
			},
		},
		{
			Name:        "tasks_delete",
			Description: "Delete a task by id.",
			InputSchema: mcpx.ObjectSchema(map[string]any{
				"id": mcpx.StringProp("Task id"),
			}, "id"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var in struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(args, &in); err != nil {
					return nil, httpx.Invalid("invalid arguments")
				}
				if err := svc.Delete(ctx, in.ID); err != nil {
					return nil, err
				}
				return map[string]string{"deleted": in.ID}, nil
			},
		},
	}
	return tools, nil
}
