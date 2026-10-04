package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leafsource/gosilite/app"
	"github.com/leafsource/gosilite/examples/tasks"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := app.FromEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := tasks.New(ctx, cfg)
	if err != nil {
		logger.Error("boot failed", "err", err)
		os.Exit(1)
	}
	defer application.Close()

	// Business MCP on stdio is for agent clients when requested.
	if cfg.MCPStdio {
		server, err := application.MCPServer(cfg)
		if err != nil {
			logger.Error("mcp setup failed", "err", err)
			os.Exit(1)
		}
		if err := server.ServeStdio(ctx, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
			logger.Error("mcp serve failed", "err", err)
			os.Exit(1)
		}
		return
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.HTTPHandler(logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}
}
