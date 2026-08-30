package chainstore

import (
	"errors"
	"fmt"
	"strings"
)

// ErrPermanent indicates a chain write failure that should not be retried.
var ErrPermanent = errors.New("chain: permanent error")

// PermanentError wraps a non-retryable submission failure.
func PermanentError(msg string) error {
	return fmt.Errorf("%w: %s", ErrPermanent, msg)
}

// IsRetryable reports whether txqueue should keep retrying this error.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrPermanent) {
		return false
	}
	msg := strings.ToLower(err.Error())
	nonRetry := []string{
		"reverted",
		"unauthorized",
		"no ledger",
		"ledger not found",
		"not configured",
	}
	for _, s := range nonRetry {
		if strings.Contains(msg, s) {
			return false
		}
	}
	retryable := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporarily unavailable",
		"waiting receipt",
		"wait mined",
		"nonce",
	}
	for _, s := range retryable {
		if strings.Contains(msg, s) {
			return true
		}
	}
	// Unknown transport errors: retry by default.
	return true
}
