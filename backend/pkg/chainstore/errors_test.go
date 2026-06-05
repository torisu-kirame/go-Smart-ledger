package chainstore

import (
	"errors"
	"testing"
)

func TestIsRetryable(t *testing.T) {
	if IsRetryable(PermanentError("reverted")) {
		t.Fatal("permanent should not retry")
	}
	if !IsRetryable(errors.New("connection refused")) {
		t.Fatal("network should retry")
	}
	if !IsRetryable(errors.New("fisco: timeout waiting receipt")) {
		t.Fatal("timeout should retry")
	}
}

func TestParseFiscoBlockNumberFromJSON(t *testing.T) {
	if n := parseFiscoBlockNumber([]byte(`"0xa"`)); n != 10 {
		t.Fatalf("hex block: %d", n)
	}
	if n := parseFiscoBlockNumber([]byte(`12`)); n != 12 {
		t.Fatalf("int block: %d", n)
	}
}
