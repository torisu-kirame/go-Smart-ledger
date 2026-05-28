package accounting

import "testing"

func TestBuildRevaluationReport(t *testing.T) {
	settings := DefaultCurrencySettings("L1")
	balances := []FCBalance{{
		AccountCode: "1002", Currency: "USD",
		ForeignBalance: "100.00", BookBalance: "720.00",
	}}
	rates := PeriodFxRates{
		Period: "2026-03", BaseCurrency: "CNY",
		Rates: []FxRateRow{{Currency: "USD", Rate: "7.50"}},
	}
	rep := BuildRevaluationReport(settings, balances, rates)
	if len(rep.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(rep.Lines))
	}
	gl, _ := ParseAmount(rep.Lines[0].GainLoss)
	// 100 * 7.5 = 750, gain 30
	if gl.Cmp(parseMust("30.00")) != 0 {
		t.Fatalf("gain expected 30.00, got %s", rep.Lines[0].GainLoss)
	}
}
