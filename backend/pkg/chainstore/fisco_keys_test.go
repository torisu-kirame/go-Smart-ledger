package chainstore

import "testing"

func TestLedgerIDFromStateKey(t *testing.T) {
	cases := []struct {
		key string
		id  string
	}{
		{"smartledger:ledger:abc", "abc"},
		{"smartledger:ledger:abc:event:1", "abc"},
		{"smartledger:ledger:abc:pending:p1", "abc"},
		{"smartledger:__ledger_ids__", fiscoGlobalLedgerID},
	}
	for _, c := range cases {
		if got := ledgerIDFromStateKey(c.key); got != c.id {
			t.Fatalf("ledgerIDFromStateKey(%q) = %q, want %q", c.key, got, c.id)
		}
	}
}

func TestIsLedgerMetaKey(t *testing.T) {
	if !isLedgerMetaKey("smartledger:ledger:abc") {
		t.Fatal("meta key")
	}
	if isLedgerMetaKey("smartledger:ledger:abc:event:1") {
		t.Fatal("event key")
	}
}
