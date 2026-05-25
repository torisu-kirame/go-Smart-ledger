package domain

import "testing"

func TestMultiRequiresAtLeastOne(t *testing.T) {
	if ValidateCreate(LedgerMulti, nil) != ErrInvalidMember {
		t.Fatal("expected ErrInvalidMember")
	}
	if ValidateCreate(LedgerMulti, []Member{{ID: "a"}}) != nil {
		t.Fatal("expected nil for single creator")
	}
	if ValidateCreate(LedgerMulti, []Member{{ID: "a"}, {ID: "b"}}) != nil {
		t.Fatal("expected nil")
	}
}
