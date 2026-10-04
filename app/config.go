package app

import (
	"fmt"
	"os"
	"strings"
)

// Config is process configuration loaded from the environment.
type Config struct {
	// HTTPAddr is the HTTP listen address, e.g. ":8080".
	HTTPAddr string
	// DBDSN is the database DSN (sqlite/mysql/postgres scheme).
	DBDSN string
	// MCPStdio enables the business MCP server on stdin/stdout when true.
	MCPStdio bool
	// ServiceName identifies the process in logs and MCP serverInfo.
	ServiceName string
	// ServiceVersion identifies the process version.
	ServiceVersion string
}

// FromEnv loads config with defaults suitable for local development.
func FromEnv() Config {
	return Config{
		HTTPAddr:       envOr("GOSILITE_HTTP_ADDR", ":8080"),
		DBDSN:          envOr("GOSILITE_DB_DSN", "sqlite://:memory:"),
		MCPStdio:       envBool("GOSILITE_MCP_STDIO"),
		ServiceName:    envOr("GOSILITE_SERVICE_NAME", "gosilite"),
		ServiceVersion: envOr("GOSILITE_SERVICE_VERSION", "0.1.0"),
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// Validate checks required fields.
func (c Config) Validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("app: HTTPAddr is required")
	}
	if c.DBDSN == "" {
		return fmt.Errorf("app: DBDSN is required")
	}
	return nil
}
