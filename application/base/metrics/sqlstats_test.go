package metrics

import (
	"testing"
	"time"
)

func TestNormalizeSQLAndObserve(t *testing.T) {
	Configure(1, 10, true)
	SQL().Observe("SELECT * FROM users WHERE id = 42", 5*time.Millisecond, nil)
	SQL().Observe("SELECT * FROM users WHERE id = 43", 10*time.Millisecond, nil)
	SQL().Observe("SELECT pg_sleep(10)", 200*time.Millisecond, nil)

	rows, slow, fw, hit, threshold, _ := SQL().Snapshot()
	if threshold != 1 {
		t.Fatalf("threshold=%d", threshold)
	}
	if len(rows) != 2 {
		t.Fatalf("templates=%d want 2", len(rows))
	}
	if rows[0].Count != 2 {
		t.Fatalf("first count=%d", rows[0].Count)
	}
	if len(slow) == 0 {
		t.Fatal("expect slow entry")
	}
	if hit == 0 || len(fw) == 0 {
		t.Fatalf("firewall hit=%d recent=%d", hit, len(fw))
	}
}

func TestNormalizeSQL(t *testing.T) {
	got := NormalizeSQL("SELECT * FROM t WHERE a = 'x' AND b = 1")
	if got != "SELECT * FROM t WHERE a = '?' AND b = ?" {
		t.Fatalf("got %q", got)
	}
}

func TestFirewallMultiStatementIgnoresLiterals(t *testing.T) {
	// 日志 body 内大量分号不应命中
	insert := `INSERT INTO "system_oper_log" ("trace_id","request_body") VALUES ('?',?) RETURNING "id"`
	// 字面量里带很多 ;
	insertWithSemiInLit := `INSERT INTO t (b) VALUES ('a;b;c; var x=1; y=2;') RETURNING "id"`
	if rule, hit := matchFirewall(insert); hit {
		t.Fatalf("plain insert should not hit, got %s", rule)
	}
	if rule, hit := matchFirewall(insertWithSemiInLit); hit {
		t.Fatalf("semis in literals should not hit, got %s", rule)
	}
	// 引号外真·多语句
	multi := `SELECT 1; DROP TABLE users`
	if rule, hit := matchFirewall(multi); !hit || rule != "multi_statement" {
		t.Fatalf("want multi_statement, got rule=%q hit=%v", rule, hit)
	}
	// 单句结尾分号允许
	if rule, hit := matchFirewall("SELECT 1;"); hit {
		t.Fatalf("trailing semicolon should not hit, got %s", rule)
	}
	// 字面量里的 0x 不应触发 dangerous_pattern
	if rule, hit := matchFirewall(`INSERT INTO t (b) VALUES ('0xdead'); `); hit {
		t.Fatalf("0x in literal should not hit, got %s", rule)
	}
	if rule, hit := matchFirewall(`SELECT 0x1`); !hit || rule != "dangerous_pattern" {
		t.Fatalf("0x outside literal want dangerous_pattern, got %q/%v", rule, hit)
	}
}
