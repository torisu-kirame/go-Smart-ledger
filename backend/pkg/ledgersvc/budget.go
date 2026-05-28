package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
)

// GetBudget returns budget for a period (empty if unset).
func (s *Service) GetBudget(ctx context.Context, ledgerID, userID, period string) (*accounting.PeriodBudget, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	period, err := accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	var b accounting.PeriodBudget
	if err := s.loadJSON(ctx, domain.LedgerBudgetKey(ledgerID, period), &b); err != nil || len(b.Lines) == 0 && b.Period == "" {
		empty := accounting.EmptyBudget(ledgerID, period)
		return &empty, nil
	}
	b.LedgerID = ledgerID
	b.Period = period
	return &b, nil
}

// PutBudget replaces budget lines for a period.
func (s *Service) PutBudget(ctx context.Context, ledgerID, userID string, budget accounting.PeriodBudget) (*accounting.PeriodBudget, error) {
	meta, err := s.getProfessionalLedger(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err := accounting.NormalizePeriod(budget.Period)
	if err != nil {
		return nil, err
	}
	budget.Period = period
	budget.LedgerID = ledgerID
	for i := range budget.Lines {
		if budget.Lines[i].ID == "" {
			id, err := snowflake.NextString()
			if err != nil {
				return nil, err
			}
			budget.Lines[i].ID = id
		}
	}
	if err := accounting.ValidateBudget(&budget); err != nil {
		return nil, err
	}
	budget.UpdatedAt = time.Now().UTC()
	if err := s.putJSON(ctx, domain.LedgerBudgetKey(ledgerID, period), ledgerID, budget); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"period": period, "lineCount": len(budget.Lines)})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventBudgetUpdated, raw)
	return &budget, nil
}

// GetBudgetAnalysis compares budget to journal actuals (F43).
func (s *Service) GetBudgetAnalysis(ctx context.Context, ledgerID, userID, period string) (*accounting.BudgetAnalysis, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	period, err := accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	budget, err := s.GetBudget(ctx, ledgerID, userID, period)
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
	rep := accounting.BuildBudgetAnalysis(*chart, *budget, journals)
	return &rep, nil
}
