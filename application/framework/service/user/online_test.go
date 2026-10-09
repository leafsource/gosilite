package user

import (
	"testing"
	"time"
)

func ts(minute int) time.Time {
	return time.Date(2026, 2, 7, 10, minute, 0, 0, time.Local)
}

func TestPageOnline(t *testing.T) {
	rows := []OnlineRow{
		{AccessHash: "h1", Account: "alice", DeptID: 11, LoginAt: ts(1)},
		{AccessHash: "h2", Account: "bob", DeptID: 12, LoginAt: ts(5)},
		{AccessHash: "h3", Account: "AlIcE", DeptID: 11, LoginAt: ts(3)},
	}
	deptNames := map[int64]string{11: "研发部", 12: "产品部"}

	// 过滤（大小写不敏感）+ 部门名
	got, total := pageOnline(rows, "ALICE", 1, 10, deptNames)
	if total != 2 {
		t.Fatalf("total want 2, got %d", total)
	}
	for _, r := range got {
		if r.Account != "alice" && r.Account != "AlIcE" {
			t.Fatalf("keyword filter wrong: %+v", r)
		}
		if r.DeptName != "研发部" {
			t.Fatalf("dept name not attached: %+v", r)
		}
	}

	// 登录时间倒序 + 分页
	got, total = pageOnline(rows, "", 1, 2, deptNames)
	if total != 3 || len(got) != 2 {
		t.Fatalf("page1: total=%d len=%d", total, len(got))
	}
	if got[0].AccessHash != "h2" || got[1].AccessHash != "h3" {
		t.Fatalf("want desc order h2,h3 got %s,%s", got[0].AccessHash, got[1].AccessHash)
	}
	got, _ = pageOnline(rows, "", 2, 2, deptNames)
	if len(got) != 1 || got[0].AccessHash != "h1" {
		t.Fatalf("page2 wrong: %+v", got)
	}

	// 空页越界
	got, _ = pageOnline(rows, "", 9, 10, deptNames)
	if len(got) != 0 {
		t.Fatalf("out-of-range page should be empty: %+v", got)
	}

	// 归一化分页参数
	got, _ = pageOnline(rows, "", 0, 0, deptNames)
	if len(got) != 3 {
		t.Fatalf("default paging wrong: %d", len(got))
	}
}
