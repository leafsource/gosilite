package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	content := "# 注释\nGO_ENV=dev\nGOSILITE_JWT_SECRET=\"abc=def\"\n\nEMPTY=\n"
	if err := os.WriteFile(filepath.Join(dir, ".env.dev"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("GOSILITE_JWT_SECRET", "from-os-env")

	vars := loadDotEnv("dev")
	if vars["GO_ENV"] != "dev" {
		t.Fatalf("GO_ENV = %q, want dev", vars["GO_ENV"])
	}
	// 系统环境变量优先于 .env 文件
	if vars["GOSILITE_JWT_SECRET"] != "from-os-env" {
		t.Fatalf("GOSILITE_JWT_SECRET = %q, want from-os-env", vars["GOSILITE_JWT_SECRET"])
	}
	if v, ok := vars["EMPTY"]; !ok || v != "" {
		t.Fatalf("EMPTY = %q ok=%v, want empty value present", v, ok)
	}
	if _, ok := vars["注释"]; ok {
		t.Fatal("注释行不应被解析")
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if vars := loadDotEnv("prod"); len(vars) == 0 {
		t.Fatal("无 .env.prod 时应返回空表而非 panic")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	conf = new(Config)
	conf.Server.EnableAccessLog = true
	conf.Server.Address = ":8888"
	conf.JWT.Secret = "yaml-default"

	applyEnvOverrides(map[string]string{
		"GOSILITE_SERVER_ADDRESS":    ":8080",
		"GOSILITE_SERVER_ACCESS_LOG": "false",
		"GOSILITE_JWT_SECRET":        "",
		"GOSILITE_STORAGE_ENDPOINT":  "http://127.0.0.1:9000",
	})

	if conf.Server.Address != ":8080" {
		t.Fatalf("Address = %q, want :8080", conf.Server.Address)
	}
	if conf.Server.EnableAccessLog {
		t.Fatal("EnableAccessLog 应被覆盖为 false")
	}
	// 空值不覆盖，回落 yaml
	if conf.JWT.Secret != "yaml-default" {
		t.Fatalf("Secret = %q, want yaml-default", conf.JWT.Secret)
	}
	if conf.Storage.Endpoint != "http://127.0.0.1:9000" {
		t.Fatalf("Endpoint = %q", conf.Storage.Endpoint)
	}
}
