package accounting

import (
	"math/big"
	"testing"
)

func TestBuildAgingReport_FIFO(t *testing.T) {
	chart := DefaultChart("L1")
	journals := []JournalEntry{
		{
			ID: "j1", Date: "2026-01-10", Period: "2026-01",
			Lines: []JournalLine{
				{AccountCode: "1122", Debit: "1000.00", Counterparty: "客户A"},
				{AccountCode: "6001", Credit: "1000.00"},
			},
		},
		{
			ID: "j2", Date: "2026-02-01", Period: "2026-02",
			Lines: []JournalLine{
				{AccountCode: "1122", Credit: "300.00", Counterparty: "客户A"},
				{AccountCode: "1002", Debit: "300.00"},
			},
		},
	}
	rep, err := BuildAgingReport(chart, journals, "2026-03-01", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(rep.Summaries))
	}
	total, _ := ParseAmount(rep.Summaries[0].Total)
	if total.Cmp(parseMust("700.00")) != 0 {
		t.Fatalf("open AR expected 700.00, got %s", rep.Summaries[0].Total)
	}
}

func parseMust(s string) *big.Int {
	v, err := ParseAmount(s)
	if err != nil {
		panic(err)
	}
	return v
}
