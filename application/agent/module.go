package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	agentRouter "gosilite/api/router/agent"
	"gosilite/application/agent/service/single_agent_draft"
	"gosilite/application/infra"
)

// Module is the agent domain plugin.
type Module struct{}

// NewModule constructs the agent plugin.
func NewModule() *Module { return &Module{} }

// Name implements application.Module.
func (m *Module) Name() string { return "agent" }

// Init implements application.Module.
func (m *Module) Init(ctx context.Context, deps *infra.AppDependencies) error {
	if deps == nil || deps.DB == nil {
		return fmt.Errorf("agent module: infra dependencies are required")
	}
	single_agent_draft.InitService(ctx, deps.DB)
	return nil
}

// RegisterRoutes implements application.Module.
func (m *Module) RegisterRoutes(r *server.Hertz) {
	agentRouter.Register(r)
}
