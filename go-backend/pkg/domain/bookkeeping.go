package domain

import "errors"

// Bookkeeping mode: simple (template entries) vs professional (COA + journals).
const (
	BookkeepingSimple        = "simple"
	BookkeepingProfessional  = "professional"
	TemplateProfessional     = "professional"
)

var ErrBookkeepingModeMismatch = errors.New("bookkeeping mode mismatch")

func NormalizeBookkeepingMode(m string) string {
	switch m {
	case BookkeepingProfessional:
		return BookkeepingProfessional
	default:
		return BookkeepingSimple
	}
}

// ResolvedBookkeepingMode returns the effective mode for a ledger (legacy → simple).
func ResolvedBookkeepingMode(meta *LedgerMeta) string {
	if meta == nil {
		return BookkeepingSimple
	}
	if meta.BookkeepingMode != "" {
		return NormalizeBookkeepingMode(meta.BookkeepingMode)
	}
	if meta.EntrySchema.TemplateID == TemplateProfessional {
		return BookkeepingProfessional
	}
	return BookkeepingSimple
}

func IsProfessionalBookkeeping(meta *LedgerMeta) bool {
	return ResolvedBookkeepingMode(meta) == BookkeepingProfessional
}

func IsSimpleBookkeeping(meta *LedgerMeta) bool {
	return !IsProfessionalBookkeeping(meta)
}

// ProfessionalEntrySchema is stored on professional ledgers (no dynamic entry columns).
func ProfessionalEntrySchema() EntrySchema {
	return EntrySchema{TemplateID: TemplateProfessional, Fields: nil}
}
