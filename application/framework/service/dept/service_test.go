package dept

import "testing"

func TestBuildTreeEmpty(t *testing.T) {
	nodes := buildTree(nil)
	if len(nodes) != 0 {
		t.Fatalf("empty tree, got %d", len(nodes))
	}
}

func TestIsDescendantUsesAncestors(t *testing.T) {
	// parent 101 在 ancestors "0,100,101" 中 → 是后代
	if !isDescendant(101, "0,100,101") {
		t.Fatal("expected 101 to be in ancestors")
	}
	if isDescendant(999, "0,100,101") {
		t.Fatal("999 not in ancestors")
	}
	if isDescendant(0, "0,100") {
		t.Fatal("0 is root, not descendant")
	}
}

// 环检测方向回归：旧实现误把「父级是自己的祖先」当环，导致未换父的更新全被拒
func TestCycleCheckDirection(t *testing.T) {
	// 不换父更新 101（父 100 的祖先链 "0" 不含 101）→ 不是环
	if isDescendant(101, "0") {
		t.Fatal("unchanged-parent update must not be treated as cycle")
	}
	// 把 100 挪到其子节点 101 下（父 101 的祖先链 "0,100" 含 100）→ 成环
	if !isDescendant(100, "0,100") {
		t.Fatal("moving dept under its own descendant must be cycle")
	}
}
