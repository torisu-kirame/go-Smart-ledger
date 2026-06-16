package ledgersvc

import (
	"context"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/accounting"
)

// GetAgingReport builds AR/AP aging (F45).
func (s *Service) GetAgingReport(ctx context.Context, ledgerID, userID, asOf, recvCSV, payCSV string) (*accounting.AgingReport, error) {
	if _, err := s.getProfessionalLedger(ctx, ledgerID, userID); err != nil {
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
	recv := splitCodes(recvCSV)
	pay := splitCodes(payCSV)
	rep, err := accounting.BuildAgingReport(*chart, journals, asOf, recv, pay)
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func splitCodes(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
