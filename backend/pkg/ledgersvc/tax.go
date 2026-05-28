package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

// ListTaxPresets returns built-in tax templates.
func (s *Service) ListTaxPresets() []accounting.BuiltinTaxPreset {
	return accounting.BuiltinTaxPresets()
}

// GetTaxTemplate returns ledger tax settings.
func (s *Service) GetTaxTemplate(ctx context.Context, ledgerID, userID string) (*accounting.TaxTemplate, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	var t accounting.TaxTemplate
	if err := s.loadJSON(ctx, domain.LedgerTaxTemplateKey(ledgerID), &t); err != nil || t.LedgerID == "" {
		def := accounting.DefaultTaxTemplate(ledgerID)
		return &def, nil
	}
	t.LedgerID = ledgerID
	return &t, nil
}

// PutTaxTemplate updates tax settings.
func (s *Service) PutTaxTemplate(ctx context.Context, ledgerID, userID string, t accounting.TaxTemplate) (*accounting.TaxTemplate, error) {
	meta, err := s.getProfessionalLedger(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	t.LedgerID = ledgerID
	if err := accounting.ValidateTaxTemplate(&t); err != nil {
		return nil, err
	}
	t.UpdatedAt = time.Now().UTC()
	if err := s.putJSON(ctx, domain.LedgerTaxTemplateKey(ledgerID), ledgerID, t); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"mode": t.Mode})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventTaxTemplateUpdated, raw)
	return &t, nil
}

// ApplyTaxPreset applies a built-in preset to the ledger template.
func (s *Service) ApplyTaxPreset(ctx context.Context, ledgerID, userID, presetID string) (*accounting.TaxTemplate, error) {
	t, err := s.GetTaxTemplate(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if !accounting.ApplyTaxPreset(t, presetID) {
		return nil, accounting.ErrInvalidTax
	}
	return s.PutTaxTemplate(ctx, ledgerID, userID, *t)
}

// GetTaxReport builds period VAT summary (F47).
func (s *Service) GetTaxReport(ctx context.Context, ledgerID, userID, period string) (*accounting.TaxReport, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	normPeriod, err := accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	tmpl, err := s.GetTaxTemplate(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	chart, err := s.GetChart(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	journals, err := s.ListJournals(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	rep := accounting.BuildTaxReport(*chart, *tmpl, journals, normPeriod)
	return &rep, nil
}
