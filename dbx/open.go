package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// DB wraps *sql.DB with the active dialect.
type DB struct {
	SQL     *sql.DB
	Dialect Dialect
	Driver  string
}

// Open connects using a DSN with an explicit scheme.
//
// Supported schemes:
//   - sqlite://path, file:path, :memory:
//   - mysql://user:pass@host:port/dbname
//   - postgres://user:pass@host:port/dbname or postgresql://
func Open(ctx context.Context, dsn string) (*DB, error) {
	driver, resolved, err := resolveDSN(dsn)
	if err != nil {
		return nil, err
	}
	dialect, err := DialectFor(driver)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driver, resolved)
	if err != nil {
		return nil, fmt.Errorf("dbx: open: %w", err)
	}
	// A plain in-memory SQLite DSN is per-connection; pin the pool to one conn
	// so schema and data stay coherent across HTTP/MCP handlers.
	if driver == "sqlite" && isMemoryDSN(resolved) {
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("dbx: ping: %w", err)
	}
	return &DB{SQL: db, Dialect: dialect, Driver: driver}, nil
}

func isMemoryDSN(resolved string) bool {
	if resolved == ":memory:" {
		return true
	}
	if !strings.HasPrefix(resolved, "file:") {
		return false
	}
	if strings.Contains(resolved, ":memory:") || strings.Contains(resolved, "mode=memory") {
		return true
	}
	return false
}

// Close closes the underlying pool.
func (d *DB) Close() error {
	if d == nil || d.SQL == nil {
		return nil
	}
	return d.SQL.Close()
}

func resolveDSN(dsn string) (driver, resolved string, err error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", "", fmt.Errorf("dbx: empty DSN")
	}
	if dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return "sqlite", dsn, nil
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("dbx: parse dsn: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "sqlite":
		path := strings.TrimPrefix(dsn, "sqlite://")
		if path == "" {
			path = ":memory:"
		}
		if path != ":memory:" && !strings.HasPrefix(path, "file:") {
			path = "file:" + path
		}
		return "sqlite", path, nil
	case "mysql":
		// go-sql-driver expects user:pass@tcp(host:port)/db
		if u.User == nil || u.Host == "" || strings.TrimPrefix(u.Path, "/") == "" {
			return "", "", fmt.Errorf("dbx: mysql DSN must be mysql://user:pass@host:port/db")
		}
		auth := u.User.Username()
		if pass, ok := u.User.Password(); ok {
			auth += ":" + pass
		}
		q := u.Query()
		if q.Get("parseTime") == "" {
			q.Set("parseTime", "true")
		}
		resolved = fmt.Sprintf("%s@tcp(%s)/%s?%s", auth, u.Host, strings.TrimPrefix(u.Path, "/"), q.Encode())
		return "mysql", resolved, nil
	case "postgres", "postgresql":
		return "postgres", dsn, nil
	default:
		return "", "", fmt.Errorf("dbx: unsupported DSN scheme %q", u.Scheme)
	}
}
