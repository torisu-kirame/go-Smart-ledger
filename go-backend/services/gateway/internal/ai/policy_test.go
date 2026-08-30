package ai

import "testing"

func TestNormalizeAndAllow(t *testing.T) {
	p, err := normalizeAPIPath("/api/v1/ledgers/{ledgerId}/events", "abc")
	if err != nil || p != "/api/v1/ledgers/abc/events" {
		t.Fatalf("got %q %v", p, err)
	}
	if err := assertAPIAllowed("GET", p); err != nil {
		t.Fatal(err)
	}
	if err := assertAPIAllowed("POST", "/api/v1/ai/chat"); err == nil {
		t.Fatal("expected block")
	}
}
