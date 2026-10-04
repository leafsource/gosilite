package app

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("GOSILITE_HTTP_ADDR", "")
	t.Setenv("GOSILITE_DB_DSN", "")
	t.Setenv("GOSILITE_MCP_STDIO", "")
	cfg := FromEnv()
	if cfg.HTTPAddr != ":8080" || cfg.DBDSN != "sqlite://:memory:" || cfg.MCPStdio {
		t.Fatalf("defaults = %+v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvBool(t *testing.T) {
	t.Setenv("GOSILITE_MCP_STDIO", "true")
	if !FromEnv().MCPStdio {
		t.Fatal("expected true")
	}
	t.Setenv("GOSILITE_MCP_STDIO", "0")
	if FromEnv().MCPStdio {
		t.Fatal("expected false")
	}
}
