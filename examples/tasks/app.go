package tasks

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/leafsource/gosilite/app"
	"github.com/leafsource/gosilite/dbx"
	"github.com/leafsource/gosilite/httpx"
	"github.com/leafsource/gosilite/mcpx"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrations returns the embedded schema migrations for this example.
func Migrations() (fs.FS, error) {
	return fs.Sub(migrationFS, "migrations")
}

// App wires database, service, HTTP routes and MCP tools for the tasks example.
type App struct {
	DB  *dbx.DB
	Svc *Service
}

// New opens the database, migrates schema and builds the service.
func New(ctx context.Context, cfg app.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	db, err := dbx.Open(ctx, cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	sub, err := Migrations()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	migrations, err := dbx.ParseMigrations(sub)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := dbx.Migrate(ctx, db.SQL, db.Dialect, migrations); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &App{
		DB:  db,
		Svc: NewService(NewSQLRepository(db)),
	}, nil
}

// Close releases the database pool.
func (a *App) Close() error {
	if a == nil || a.DB == nil {
		return nil
	}
	return a.DB.Close()
}

// HTTPHandler returns the full HTTP stack for this example.
func (a *App) HTTPHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	NewHandler(a.Svc).Routes(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.RespondOK(w, map[string]string{"status": "ok"})
	})
	return httpx.Chain(mux, httpx.Recover, httpx.RequestID, httpx.Logging(logger))
}

// MCPServer exposes business tools (not system tools) for agent clients.
func (a *App) MCPServer(cfg app.Config) (*mcpx.Server, error) {
	server := mcpx.NewServer(cfg.ServiceName, cfg.ServiceVersion, nil)
	tools, err := MCPTools(a.Svc)
	if err != nil {
		return nil, err
	}
	for _, t := range tools {
		if err := server.Register(t); err != nil {
			return nil, fmt.Errorf("register %s: %w", t.Name, err)
		}
	}
	return server, nil
}
