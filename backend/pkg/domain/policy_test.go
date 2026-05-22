package domain

import "testing"

func TestMultiRequiresTwo(t *testing.T) {
	if ValidateCreate(LedgerMulti, []Member{{ID: "a"}}) != ErrMultiNeedsTwo {
		t.Fatal("expected ErrMultiNeedsTwo")
	}
	if ValidateCreate(LedgerMulti, []Member{{ID: "a"}, {ID: "b"}}) != nil {
		t.Fatal("expected nil")
	}
}
