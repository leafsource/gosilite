package dbx

import (
	"fmt"
	"strings"
)

// Dialect abstracts small SQL differences across supported databases.
type Dialect interface {
	Name() string
	Placeholder(n int) string
	QuoteIdent(ident string) string
}

// DialectFor returns a dialect for a driver name.
func DialectFor(driver string) (Dialect, error) {
	switch driver {
	case "sqlite":
		return sqliteDialect{}, nil
	case "mysql":
		return mysqlDialect{}, nil
	case "postgres":
		return postgresDialect{}, nil
	default:
		return nil, fmt.Errorf("dbx: unsupported driver %q", driver)
	}
}

type sqliteDialect struct{}

func (sqliteDialect) Name() string { return "sqlite" }
func (sqliteDialect) Placeholder(int) string {
	return "?"
}
func (sqliteDialect) QuoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

type mysqlDialect struct{}

func (mysqlDialect) Name() string { return "mysql" }
func (mysqlDialect) Placeholder(int) string {
	return "?"
}
func (mysqlDialect) QuoteIdent(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}

type postgresDialect struct{}

func (postgresDialect) Name() string { return "postgres" }
func (postgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}
func (postgresDialect) QuoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
