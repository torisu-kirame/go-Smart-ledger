package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

// GetCurrencySettings returns FX settings or defaults.
func (s *Service) GetCurrencySettings(ctx context.Context, ledgerID, userID string) (*accounting.CurrencySettings, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	var cs accounting.CurrencySettings
	if err := s.loadJSON(ctx, domain.LedgerCurrencyKey(ledgerID), &cs); err != nil || cs.BaseCurrency == "" {
		def := accounting.DefaultCurrencySettings(ledgerID)
		return &def, nil
	}
	cs.LedgerID = ledgerID
	return &cs, nil
}

// PutCurrencySettings updates FX settings.
func (s *Service) PutCurrencySettings(ctx context.Context, ledgerID, userID string, cs accounting.CurrencySettings) (*accounting.CurrencySettings, error) {
	meta, err := s.getProfessionalLedger(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	cs.LedgerID = ledgerID
	if err := accounting.ValidateCurrencySettings(&cs); err != nil {
		return nil, err
	}
	cs.UpdatedAt = time.Now().UTC()
	if err := s.putJSON(ctx, domain.LedgerCurrencyKey(ledgerID), ledgerID, cs); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"baseCurrency": cs.BaseCurrency})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventCurrencySettingsUpdated, raw)
	return &cs, nil
}

// GetPeriodFxRates returns closing rates for a period.
func (s *Service) GetPeriodFxRates(ctx context.Context, ledgerID, userID, period string) (*accounting.PeriodFxRates, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	period, err := accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	settings, err := s.GetCurrencySettings(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	var rates accounting.PeriodFxRates
	if err := s.loadJSON(ctx, domain.LedgerFxRatesKey(ledgerID, period), &rates); err != nil {
		return &accounting.PeriodFxRates{
			LedgerID: ledgerID, Period: period, BaseCurrency: settings.BaseCurrency, Rates: nil,
		}, nil
	}
	rates.LedgerID = ledgerID
	rates.Period = period
	return &rates, nil
}

// PutPeriodFxRates saves closing rates.
func (s *Service) PutPeriodFxRates(ctx context.Context, ledgerID, userID string, rates accounting.PeriodFxRates) (*accounting.PeriodFxRates, error) {
	meta, err := s.getProfessionalLedger(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	settings, err := s.GetCurrencySettings(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err := accounting.NormalizePeriod(rates.Period)
	if err != nil {
		return nil, err
	}
	rates.Period = period
	rates.LedgerID = ledgerID
	rates.BaseCurrency = settings.BaseCurrency
	if err := accounting.ValidatePeriodFxRates(&rates); err != nil {
		return nil, err
	}
	rates.UpdatedAt = time.Now().UTC()
	if err := s.putJSON(ctx, domain.LedgerFxRatesKey(ledgerID, period), ledgerID, rates); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"period": period, "rates": len(rates.Rates)})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventFxRatesUpdated, raw)
	return &rates, nil
}

// GetFCBalances lists foreign currency balances from journals.
func (s *Service) GetFCBalances(ctx context.Context, ledgerID, userID string) ([]accounting.FCBalance, error) {
	settings, err := s.GetCurrencySettings(ctx, ledgerID, userID)
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
	return accounting.BuildFCBalances(*chart, *settings, journals), nil
}

// GetRevaluationReport runs period-end FX revaluation (F46).
func (s *Service) GetRevaluationReport(ctx context.Context, ledgerID, userID, period string) (*accounting.RevaluationReport, error) {
	settings, err := s.GetCurrencySettings(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err = accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	rates, err := s.GetPeriodFxRates(ctx, ledgerID, userID, period)
	if err != nil {
		return nil, err
	}
	if len(rates.Rates) == 0 {
		return &accounting.RevaluationReport{
			Period: period, BaseCurrency: settings.BaseCurrency, GainLossAccount: settings.GainLossAccount,
		}, nil
	}
	balances, err := s.GetFCBalances(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	rep := accounting.BuildRevaluationReport(*settings, balances, *rates)
	return &rep, nil
}
