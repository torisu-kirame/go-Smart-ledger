package chainstore

import "testing"

func TestParseFiscoQuery(t *testing.T) {
	eq, err := parseFiscoQuery(
		`SELECT key, value FROM world_state WHERE key = ?`,
		[]interface{}{"smartledger:ledger:x"},
	)
	if err != nil || eq.eqKey != "smartledger:ledger:x" {
		t.Fatalf("eq: %+v %v", eq, err)
	}

	like, err := parseFiscoQuery(
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		[]interface{}{"smartledger:ledger:a:event:%"},
	)
	if err != nil || like.likePrefix != "smartledger:ledger:a:event:%" {
		t.Fatalf("like: %+v %v", like, err)
	}

	list, err := parseFiscoQuery(
		`SELECT key, value FROM world_state WHERE key LIKE ? AND key NOT LIKE ? ORDER BY key`,
		[]interface{}{"smartledger:ledger:%", "smartledger:ledger:%:event:%"},
	)
	if err != nil || list.likePrefix == "" || list.notLike == "" {
		t.Fatalf("list: %+v %v", list, err)
	}

	suf, err := parseFiscoQuery(
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		[]interface{}{"%:invite:user1"},
	)
	if err != nil || suf.suffixMatch != "%:invite:user1" {
		t.Fatalf("suffix: %+v %v", suf, err)
	}
}

func TestLedgerIDFromLikePrefix(t *testing.T) {
	if got := ledgerIDFromLikePrefix("smartledger:ledger:abc:event:%"); got != "abc" {
		t.Fatalf("got %q", got)
	}
}
