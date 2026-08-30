package domain

import "errors"

// Bookkeeping mode: only simple (template / multi-table sheets) is supported.
const (
	BookkeepingSimple = "simple"
	// Deprecated: professional mode removed; kept so old payloads normalize to simple.
	BookkeepingProfessional = "professional"
	TemplateProfessional    = "professional"
)

// ErrBookkeepingModeMismatch is retained for HTTP mapping compatibility.
var ErrBookkeepingModeMismatch = errors.New("bookkeeping mode mismatch")

func NormalizeBookkeepingMode(m string) string {
	_ = m
	return BookkeepingSimple
}

// ResolvedBookkeepingMode always returns simple.
func ResolvedBookkeepingMode(meta *LedgerMeta) string {
	_ = meta
	return BookkeepingSimple
}

func IsProfessionalBookkeeping(meta *LedgerMeta) bool {
	_ = meta
	return false
}

func IsSimpleBookkeeping(meta *LedgerMeta) bool {
	_ = meta
	return true
}
