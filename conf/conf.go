package conf

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env string

	Server        Server        `yaml:"server"`
	JWT           JWT           `yaml:"jwt"`
	Database      Database      `yaml:"database"`
	Redis         Redis         `yaml:"redis"`
	Storage       Storage       `yaml:"storage"`
	Monitor       Monitor       `yaml:"monitor"`
	Docs          Docs          `yaml:"docs"`
	Scheduler     Scheduler     `yaml:"scheduler"`
	LoginSecurity LoginSecurity `yaml:"login_security"`
}

// Docs 开发文档静态站（Hugo 产物）托管配置。
type Docs struct {
	Dir string `yaml:"dir"`
}

type Scheduler struct {
	LogRetentionDays int `yaml:"log_retention_days" validate:"min=1"`
}

type LoginSecurity struct {
	MaxFailures          int `yaml:"max_failures" validate:"min=1,max=9"`
	FailureWindowMinutes int `yaml:"failure_window_minutes" validate:"min=1"`
	LockMinutes          int `yaml:"lock_minutes" validate:"min=1"`
}

// Monitor 数据监控：指标出口与采集开关（业务热路径仅 O(1) 打点）。
type Monitor struct {
	EnableMetrics       bool   `yaml:"enable_metrics"`
	MetricsAddr         string `yaml:"metrics_addr"`
	MetricsPath         string `yaml:"metrics_path"`
	SQLSlowMs           int    `yaml:"sql_slow_ms"`
	SQLMaxTemplates     int    `yaml:"sql_max_templates"`
	FirewallEnable      bool   `yaml:"firewall_enable"`
	GormPostgresMetrics bool   `yaml:"gorm_postgres_metrics"`
}

type Database struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

type Server struct {
	Node            int64  `yaml:"node"`
	Address         string `yaml:"address"`
	EnablePprof     bool   `yaml:"enable_pprof"`
	EnableGzip      bool   `yaml:"enable_gzip"`
	EnableAccessLog bool   `yaml:"enable_access_log"`
	LogLevel        string `yaml:"log_level"`
	LogFileName     string `yaml:"log_file_name"`
	LogMaxSize      int    `yaml:"log_max_size"`
	LogMaxBackups   int    `yaml:"log_max_backups"`
	LogMaxAge       int    `yaml:"log_max_age"`
}

type JWT struct {
	Issuer string `yaml:"issuer"`
	Secret string `yaml:"secret"`
}

type Storage struct {
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Endpoint   string `yaml:"endpoint"`
	Region     string `yaml:"region"`
	BucketName string `yaml:"bucket_name"`
	Domain     string `yaml:"domain"`
}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {
	confFileRelPath := filepath.Join("conf", "conf.yaml")
	content, err := os.ReadFile(confFileRelPath)
	if err != nil {
		panic(err)
	}

	conf = new(Config)
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}

	envVars := loadDotEnv(GetEnv())
	applyEnvOverrides(envVars)

	if err := validator.Validate(conf); err != nil {
		hlog.Error("validate config error - %v", err)
		panic(err)
	}

	if GetEnv() == "prod" {
		// 生产密钥只走环境注入，禁止落到 yaml（conf.yaml 里是本地开发默认值）。
		for _, key := range []string{
			"GOSILITE_JWT_SECRET",
			"GOSILITE_DATABASE_DSN",
			"GOSILITE_STORAGE_ACCESS_KEY",
			"GOSILITE_STORAGE_SECRET_KEY",
		} {
			if envVars[key] == "" {
				panic("prod 环境缺少密钥注入: " + key + "（见仓库根 .env.example）")
			}
		}
	}

	conf.Env = GetEnv()
}

// loadDotEnv 读取仓库根 .env.<env>（KEY=VALUE，# 为注释），再并入系统环境变量；
// 同名时系统环境变量优先。返回合并后的键值表。
func loadDotEnv(env string) map[string]string {
	vars := make(map[string]string)
	if b, err := os.ReadFile(".env." + env); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			vars[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			vars[k] = v
		}
	}
	return vars
}

// applyEnvOverrides 用环境变量覆盖 yaml 中的敏感项与环境差异项（空值不覆盖）。
func applyEnvOverrides(vars map[string]string) {
	set := func(key string, dst *string) {
		if v := vars[key]; v != "" {
			*dst = v
		}
	}
	set("GOSILITE_SERVER_ADDRESS", &conf.Server.Address)
	if v := vars["GOSILITE_SERVER_ACCESS_LOG"]; v != "" {
		conf.Server.EnableAccessLog = v == "true"
	}
	set("GOSILITE_JWT_SECRET", &conf.JWT.Secret)
	set("GOSILITE_DATABASE_DSN", &conf.Database.DSN)
	set("GOSILITE_REDIS_PASSWORD", &conf.Redis.Password)
	set("GOSILITE_STORAGE_ACCESS_KEY", &conf.Storage.AccessKey)
	set("GOSILITE_STORAGE_SECRET_KEY", &conf.Storage.SecretKey)
	set("GOSILITE_STORAGE_ENDPOINT", &conf.Storage.Endpoint)
	set("GOSILITE_STORAGE_DOMAIN", &conf.Storage.Domain)
	set("GOSILITE_DOCS_DIR", &conf.Docs.Dir)
	if v := vars["GOSILITE_SCHEDULER_LOG_RETENTION_DAYS"]; v != "" {
		days, err := strconv.Atoi(v)
		if err != nil || days < 1 {
			panic("GOSILITE_SCHEDULER_LOG_RETENTION_DAYS must be a positive integer")
		}
		conf.Scheduler.LogRetentionDays = days
	}
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}

func LogLevel() hlog.Level {
	level := GetConf().Server.LogLevel
	switch level {
	case "trace":
		return hlog.LevelTrace
	case "debug":
		return hlog.LevelDebug
	case "info":
		return hlog.LevelInfo
	case "notice":
		return hlog.LevelNotice
	case "warn":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	case "fatal":
		return hlog.LevelFatal
	default:
		return hlog.LevelInfo
	}
}
