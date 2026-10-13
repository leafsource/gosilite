package metrics

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// SQL 监控 / 防火墙：查询后 O(1) 内存聚合，有界存储，不落库。

var (
	numRe    = regexp.MustCompile(`\b\d+\b`)
	quoteRe  = regexp.MustCompile(`'(?:[^']|'')*'`)
	spaceRe  = regexp.MustCompile(`\s+`)
	sqlStore = &SQLStore{}
)

// SQLStore 模板聚合 + 慢 SQL 环 + 防火墙命中环。
type SQLStore struct {
	mu           sync.RWMutex
	templates    map[string]*sqlAgg
	slow         []SlowSQL
	fwHits       []FwHit
	maxTemplates int
	maxSlow      int
	maxFw        int
	slowMs       time.Duration
	fwEnabled    bool
	oversized    bool
}

type sqlAgg struct {
	Count      int64
	TotalMs    int64
	MaxMs      int64
	ErrorCount int64
	LastMs     int64
}

// SlowSQL 慢查询样本（环形保留）。
type SlowSQL struct {
	SQL    string `json:"sql"`
	CostMs int64  `json:"costMs"`
	At     int64  `json:"at"`
	Err    bool   `json:"err"`
}

// FwHit 防火墙命中样本（观察模式，不拦截）。
type FwHit struct {
	Rule string `json:"rule"`
	SQL  string `json:"sql"`
	At   int64  `json:"at"`
}

// SQLAggRow 模板聚合行。
type SQLAggRow struct {
	SQL        string `json:"sql"`
	Count      int64  `json:"count"`
	AvgMs      int64  `json:"avgMs"`
	MaxMs      int64  `json:"maxMs"`
	TotalMs    int64  `json:"totalMs"`
	ErrorCount int64  `json:"errorCount"`
	LastMs     int64  `json:"lastMs"`
}

// Configure 启动时由 conf 调用一次。
func Configure(slowMs, maxTemplates int, firewall bool) {
	if slowMs <= 0 {
		slowMs = 200
	}
	if maxTemplates <= 0 {
		maxTemplates = 200
	}
	sqlStore.mu.Lock()
	defer sqlStore.mu.Unlock()
	sqlStore.slowMs = time.Duration(slowMs) * time.Millisecond
	sqlStore.maxTemplates = maxTemplates
	sqlStore.maxSlow = 50
	sqlStore.maxFw = 50
	sqlStore.fwEnabled = firewall
	if sqlStore.templates == nil {
		sqlStore.templates = make(map[string]*sqlAgg, maxTemplates)
	}
}

func init() {
	Configure(0, 0, false)
}

// SQL 返回全局 SQL 统计存储。
func SQL() *SQLStore { return sqlStore }

// Observe 记录一次执行（GORM logger Trace 调用）。
func (s *SQLStore) Observe(rawSQL string, cost time.Duration, err error) {
	if rawSQL == "" {
		return
	}
	tpl := NormalizeSQL(rawSQL)
	ms := cost.Milliseconds()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.templates == nil {
		s.templates = make(map[string]*sqlAgg, s.maxTemplates)
	}
	agg, ok := s.templates[tpl]
	if !ok {
		if len(s.templates) >= s.maxTemplates {
			s.oversized = true
			// 满则丢弃新模板，保序不影响已有聚合
			return
		}
		agg = &sqlAgg{}
		s.templates[tpl] = agg
	}
	agg.Count++
	agg.TotalMs += ms
	if ms > agg.MaxMs {
		agg.MaxMs = ms
	}
	agg.LastMs = ms
	if err != nil {
		agg.ErrorCount++
	}

	// 防火墙观察（关闭时零规则成本）
	if s.fwEnabled {
		if rule, hit := matchFirewall(rawSQL); hit {
			s.fwHits = appendFw(s.fwHits, FwHit{Rule: rule, SQL: tpl, At: time.Now().Unix()}, s.maxFw)
		}
	}

	// 慢 SQL 环
	if cost >= s.slowMs {
		s.slow = appendSlow(s.slow, SlowSQL{SQL: tpl, CostMs: ms, At: time.Now().Unix(), Err: err != nil}, s.maxSlow)
	}
}

// Snapshot 返回模板聚合（按次数倒序）与慢 SQL、防火墙命中。
func (s *SQLStore) Snapshot() (rows []SQLAggRow, slow []SlowSQL, fw []FwHit, fwCount int64, slowMs int64, oversized bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows = make([]SQLAggRow, 0, len(s.templates))
	for tpl, a := range s.templates {
		avg := int64(0)
		if a.Count > 0 {
			avg = a.TotalMs / a.Count
		}
		rows = append(rows, SQLAggRow{
			SQL: tpl, Count: a.Count, AvgMs: avg, MaxMs: a.MaxMs,
			TotalMs: a.TotalMs, ErrorCount: a.ErrorCount, LastMs: a.LastMs,
		})
	}
	// 插入排序量小，直接 sort
	sortSQLRows(rows)
	slow = append([]SlowSQL(nil), s.slow...)
	fw = append([]FwHit(nil), s.fwHits...)
	fwCount = int64(len(s.fwHits))
	slowMs = s.slowMs.Milliseconds()
	oversized = s.oversized
	return
}

func sortSQLRows(rows []SQLAggRow) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].Count > rows[j-1].Count; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}

func appendSlow(buf []SlowSQL, v SlowSQL, max int) []SlowSQL {
	buf = append(buf, v)
	if len(buf) > max {
		buf = buf[len(buf)-max:]
	}
	return buf
}

func appendFw(buf []FwHit, v FwHit, max int) []FwHit {
	buf = append(buf, v)
	if len(buf) > max {
		buf = buf[len(buf)-max:]
	}
	return buf
}

// NormalizeSQL 去字面量/数字/多余空白，限长；作聚合键与展示。
func NormalizeSQL(q string) string {
	q = quoteRe.ReplaceAllString(q, "'?'")
	q = numRe.ReplaceAllString(q, "?")
	q = spaceRe.ReplaceAllString(q, " ")
	q = strings.TrimSpace(q)
	if len(q) > 480 {
		q = q[:480] + "…"
	}
	return q
}

// matchFirewall 廉价启发式（观察模式，不阻断业务）。
func matchFirewall(sql string) (string, bool) {
	low := strings.ToLower(sql)
	// 多语句：仅统计引号外分号；字面量里的 ;（JSON/JS/CSS 等）不算
	if isMultiStatement(sql) {
		return "multi_statement", true
	}
	// 危险片段同样避开字符串字面量，避免 request_body 等误报
	stripped := stripSQLLiterals(low)
	for _, p := range []string{"pg_sleep", "xp_cmdshell", "benchmark(", "load_file(", "0x"} {
		if strings.Contains(stripped, p) {
			return "dangerous_pattern", true
		}
	}
	if strings.Contains(stripped, "union select") {
		return "union_injection", true
	}
	return "", false
}

// isMultiStatement 引号外是否还有「中间分号」（分号后仍有非空白语句内容）。
// 单句结尾允许一个 `;`；`'a;b;c'` 这类字面量内分号忽略。
func isMultiStatement(sql string) bool {
	seenSep := false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '\'':
			j := skipQuoted(sql, i, '\'')
			if j > i {
				i = j - 1
			}
		case c == '"':
			j := skipQuoted(sql, i, '"')
			if j > i {
				i = j - 1
			}
		case c == '$':
			if end := skipDollarQuote(sql, i); end > i {
				i = end - 1
			}
		case c == ';':
			// 分号后还有非空白 → 多语句；结尾单独 ; 不算
			if strings.TrimSpace(sql[i+1:]) != "" {
				seenSep = true
			}
		}
	}
	return seenSep
}

// stripSQLLiterals 引号内内容替换为空串，只保留结构字符，避免 body 里 0x 等误报。
func stripSQLLiterals(sql string) string {
	if sql == "" {
		return sql
	}
	var b strings.Builder
	b.Grow(len(sql))
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '\'':
			j := skipQuoted(sql, i, '\'')
			b.WriteString(`''`)
			i = j - 1
		case c == '"':
			j := skipQuoted(sql, i, '"')
			b.WriteString(`""`)
			i = j - 1
		case c == '$':
			if end := skipDollarQuote(sql, i); end > i {
				b.WriteString(`$x$`)
				i = end - 1
			} else {
				b.WriteByte(c)
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// skipQuoted 返回结束引号后的下标（含闭合）；未闭合则返回 len。
func skipQuoted(s string, start int, quote byte) int {
	for j := start + 1; j < len(s); j++ {
		if s[j] != quote {
			continue
		}
		// '' 或 "" 转义
		if j+1 < len(s) && s[j+1] == quote {
			j++
			continue
		}
		return j + 1
	}
	return len(s)
}

// skipDollarQuote 支持 $tag$ ... $tag$；不匹配则返回 start。
func skipDollarQuote(s string, start int) int {
	end := strings.IndexByte(s[start+1:], '$')
	if end < 0 {
		return start
	}
	tag := s[start : start+1+end+1] // $tag$
	rest := s[start+len(tag):]
	if idx := strings.Index(rest, tag); idx >= 0 {
		return start + len(tag) + idx + len(tag)
	}
	return len(s)
}
