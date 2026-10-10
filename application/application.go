package application

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gosilite/application/framework"
	"gosilite/application/infra"
	"gosilite/pkg/ctxcache"
)

// App is the bootstrapped application with initialized modules.
type App struct {
	deps    *infra.AppDependencies
	modules []Module
}

// builtinModules is the single place to register domain plugins.
func builtinModules() []Module {
	return []Module{
		framework.NewModule(),
	}
}

// Bootstrap initializes infrastructure and every registered module.
func Bootstrap(ctx context.Context) (*App, error) {
	ctx = ctxcache.Init(ctx)

	deps, err := infra.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("init infra failed: %w", err)
	}

	mods := builtinModules()
	for _, m := range mods {
		if err := m.Init(ctx, deps); err != nil {
			return nil, fmt.Errorf("init module %s failed: %w", m.Name(), err)
		}
	}
	return &App{deps: deps, modules: mods}, nil
}

// MountRoutes registers HTTP routes for all modules.
func (a *App) MountRoutes(r *server.Hertz) {
	for _, m := range a.modules {
		m.RegisterRoutes(r)
	}
}

// Deps exposes infrastructure dependencies (DB, cache, storage).
func (a *App) Deps() *infra.AppDependencies {
	return a.deps
}

// Modules returns registered modules (read-only use).
func (a *App) Modules() []Module {
	return a.modules
}
