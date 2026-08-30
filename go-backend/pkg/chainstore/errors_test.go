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
	if !IsRetryable(errors.New("timeout waiting receipt")) {
		t.Fatal("timeout should retry")
	}
}
