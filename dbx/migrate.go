package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// Migration is one versioned schema change.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// ParseMigrations reads ordered `NNNN_name.sql` files from fsys.
func ParseMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("dbx: read migrations: %w", err)
	}
	var out []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("dbx: migration %q must be NNNN_name.sql", e.Name())
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("dbx: migration %q version: %w", e.Name(), err)
		}
		raw, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, Migration{Version: version, Name: parts[1], SQL: extractUp(string(raw))})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	for i := 1; i < len(out); i++ {
		if out[i].Version == out[i-1].Version {
			return nil, fmt.Errorf("dbx: duplicate migration version %d", out[i].Version)
		}
	}
	return out, nil
}

// extractUp keeps only the up section when a `-- +down` marker line exists.
func extractUp(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- +down" {
			break
		}
		if trimmed == "-- +up" {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// Migrate applies pending migrations in order.
// Statement boundaries are split so multi-statement files work on MySQL/PostgreSQL/SQLite.
func Migrate(ctx context.Context, db *sql.DB, dialect Dialect, migrations []Migration) (applied int, err error) {
	if dialect == nil {
		return 0, fmt.Errorf("dbx: dialect is required")
	}
	if err := ensureMigrationTable(ctx, db, dialect); err != nil {
		return 0, err
	}
	done, err := appliedVersions(ctx, db)
	if err != nil {
		return 0, err
	}
	for _, m := range migrations {
		if done[m.Version] {
			continue
		}
		if err := applyOne(ctx, db, dialect, m); err != nil {
			return applied, err
		}
		applied++
	}
	return applied, nil
}

func applyOne(ctx context.Context, db *sql.DB, dialect Dialect, m Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, stmt := range splitStatements(m.SQL) {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("dbx: migrate %04d_%s: %w", m.Version, m.Name, err)
		}
	}
	insert := fmt.Sprintf(
		`INSERT INTO schema_migrations (version, name, checksum) VALUES (%s, %s, %s)`,
		dialect.Placeholder(1), dialect.Placeholder(2), dialect.Placeholder(3),
	)
	if _, err := tx.ExecContext(ctx, insert, m.Version, m.Name, checksum(m.SQL)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("dbx: record migration %d: %w", m.Version, err)
	}
	return tx.Commit()
}

// splitStatements splits SQL on semicolon boundaries while ignoring quoted strings.
func splitStatements(sqlText string) []string {
	var (
		out       []string
		b         strings.Builder
		inSingle  bool
		inDouble  bool
		inLineCom bool
		inBlkCom  bool
	)
	flush := func() {
		stmt := strings.TrimSpace(b.String())
		b.Reset()
		if stmt != "" {
			out = append(out, stmt)
		}
	}
	runes := []rune(sqlText)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		next := func() rune {
			if i+1 < len(runes) {
				return runes[i+1]
			}
			return 0
		}
		if inLineCom {
			if c == '\n' {
				inLineCom = false
				b.WriteRune(c)
			}
			continue
		}
		if inBlkCom {
			if c == '*' && next() == '/' {
				inBlkCom = false
				i++
			}
			continue
		}
		if !inSingle && !inDouble {
			if c == '-' && next() == '-' {
				inLineCom = true
				i++
				continue
			}
			if c == '/' && next() == '*' {
				inBlkCom = true
				i++
				continue
			}
			if c == ';' {
				flush()
				continue
			}
		}
		if c == '\'' && !inDouble {
			inSingle = !inSingle
			b.WriteRune(c)
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			b.WriteRune(c)
			continue
		}
		b.WriteRune(c)
	}
	flush()
	return out
}

func ensureMigrationTable(ctx context.Context, db *sql.DB, dialect Dialect) error {
	_ = dialect
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  checksum TEXT NOT NULL
)`)
	return err
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func checksum(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())
}
