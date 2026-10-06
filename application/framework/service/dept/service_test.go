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
