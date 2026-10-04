package dbx

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDialectPlaceholder(t *testing.T) {
	tests := []struct {
		driver string
		n      int
		want   string
	}{
		{"sqlite", 1, "?"},
		{"mysql", 2, "?"},
		{"postgres", 1, "$1"},
		{"postgres", 3, "$3"},
	}
	for _, tt := range tests {
		d, err := DialectFor(tt.driver)
		if err != nil {
			t.Fatal(err)
		}
		if got := d.Placeholder(tt.n); got != tt.want {
			t.Fatalf("%s placeholder(%d) = %q, want %q", tt.driver, tt.n, got, tt.want)
		}
	}
}

func TestDialectQuoteIdent(t *testing.T) {
	d, _ := DialectFor("mysql")
	if got := d.QuoteIdent("order"); got != "`order`" {
		t.Fatalf("mysql quote = %q", got)
	}
	d, _ = DialectFor("postgres")
	if got := d.QuoteIdent(`we"ird`); got != `"we""ird"` {
		t.Fatalf("postgres quote = %q", got)
	}
}

func TestResolveDSN(t *testing.T) {
	driver, resolved, err := resolveDSN("sqlite://:memory:")
	if err != nil || driver != "sqlite" || resolved != ":memory:" {
		t.Fatalf("sqlite dsn = %q %q %v", driver, resolved, err)
	}
	driver, resolved, err = resolveDSN("mysql://user:pass@localhost:3306/app")
	if err != nil || driver != "mysql" {
		t.Fatalf("mysql dsn = %q %q %v", driver, resolved, err)
	}
	if resolved != "user:pass@tcp(localhost:3306)/app?parseTime=true" {
		t.Fatalf("mysql resolved = %q", resolved)
	}
	driver, _, err = resolveDSN("postgres://u:p@localhost:5432/app")
	if err != nil || driver != "postgres" {
		t.Fatalf("postgres dsn err = %v", err)
	}
	if _, _, err := resolveDSN("redis://x"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestMigrateSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	fsys := fstest.MapFS{
		"0001_init.sql":  {Data: []byte("CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT NOT NULL);")},
		"0002_add.sql":   {Data: []byte("-- +up\nALTER TABLE tasks ADD COLUMN status TEXT NOT NULL DEFAULT 'todo';\n-- +down\n")},
		"0003_index.sql": {Data: []byte("CREATE INDEX idx_tasks_status ON tasks(status);")},
	}
	migrations, err := ParseMigrations(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 3 {
		t.Fatalf("parsed %d migrations", len(migrations))
	}
	applied, err := Migrate(ctx, db.SQL, db.Dialect, migrations)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 3 {
		t.Fatalf("applied = %d", applied)
	}
	applied, err = Migrate(ctx, db.SQL, db.Dialect, migrations)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Fatalf("second migrate applied = %d", applied)
	}
	if _, err := db.SQL.ExecContext(ctx, `INSERT INTO tasks (title, status) VALUES ('x', 'todo')`); err != nil {
		t.Fatal(err)
	}
}

func TestSplitStatementsAndPostgresPlaceholders(t *testing.T) {
	stmts := splitStatements("CREATE TABLE t (id int); -- comment\nCREATE INDEX i ON t(id);")
	if len(stmts) != 2 {
		t.Fatalf("stmts=%d %#v", len(stmts), stmts)
	}
	d, _ := DialectFor("postgres")
	insert := fmt.Sprintf(
		`INSERT INTO schema_migrations (version, name, checksum) VALUES (%s, %s, %s)`,
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3),
	)
	if !strings.Contains(insert, "$1") || strings.Contains(insert, "?") {
		t.Fatalf("insert = %q", insert)
	}
}

func TestMemorySQLiteIsSingleConn(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db.SQL.Stats().MaxOpenConnections != 1 {
		t.Fatalf("MaxOpenConnections = %d", db.SQL.Stats().MaxOpenConnections)
	}
	fsys := fstest.MapFS{
		"0001_init.sql": {Data: []byte("CREATE TABLE t (id TEXT PRIMARY KEY);")},
	}
	migrations, err := ParseMigrations(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(ctx, db.SQL, db.Dialect, migrations); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL.ExecContext(ctx, `INSERT INTO t (id) VALUES ('a')`); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM t`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count = %d", n)
	}
}
