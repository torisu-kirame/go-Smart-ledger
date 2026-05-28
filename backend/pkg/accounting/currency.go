package accounting

import (
	"math/big"
	"strings"
	"time"
)

const defaultBaseCurrency = "CNY"

// DefaultCurrencySettings for a ledger.
func DefaultCurrencySettings(ledgerID string) CurrencySettings {
	return CurrencySettings{
		LedgerID:        ledgerID,
		BaseCurrency:    defaultBaseCurrency,
		Currencies:      []string{defaultBaseCurrency, "USD", "EUR", "HKD"},
		GainLossAccount: "6603",
		UpdatedAt:       time.Now().UTC(),
	}
}

// ValidateCurrencySettings checks settings.
func ValidateCurrencySettings(s *CurrencySettings) error {
	if s == nil || strings.TrimSpace(s.BaseCurrency) == "" {
		return ErrInvalidCurrency
	}
	s.BaseCurrency = strings.ToUpper(strings.TrimSpace(s.BaseCurrency))
	if strings.TrimSpace(s.GainLossAccount) == "" {
		s.GainLossAccount = "6603"
	}
	seen := map[string]bool{s.BaseCurrency: true}
	var list []string
	for _, c := range s.Currencies {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		list = append(list, c)
	}
	if !seen[s.BaseCurrency] {
		list = append([]string{s.BaseCurrency}, list...)
	}
	s.Currencies = list
	return nil
}

// ValidatePeriodFxRates checks closing rates.
func ValidatePeriodFxRates(p *PeriodFxRates) error {
	if p == nil {
		return ErrInvalidFxRates
	}
	if _, err := NormalizePeriod(p.Period); err != nil {
		return ErrInvalidFxRates
	}
	p.BaseCurrency = strings.ToUpper(strings.TrimSpace(p.BaseCurrency))
	for i, r := range p.Rates {
		c := strings.ToUpper(strings.TrimSpace(r.Currency))
		if c == "" || c == p.BaseCurrency {
			return ErrInvalidFxRates
		}
		rate, err := ParseAmount(r.Rate)
		if err != nil || rate.Sign() <= 0 {
			return ErrInvalidFxRates
		}
		p.Rates[i].Currency = c
	}
	return nil
}

// BuildFCBalances aggregates foreign balances from journals.
func BuildFCBalances(chart ChartOfAccounts, settings CurrencySettings, journals []JournalEntry) []FCBalance {
	base := strings.ToUpper(settings.BaseCurrency)
	idx := AccountIndex(chart)
	type balKey struct{ code, cur string }
	fcBal := make(map[balKey]*big.Int)
	bookBal := make(map[balKey]*big.Int)

	for _, j := range journals {
		for _, ln := range j.Lines {
			cur := strings.ToUpper(strings.TrimSpace(ln.Currency))
			if cur == "" || cur == base {
				continue
			}
			code := strings.TrimSpace(ln.AccountCode)
			acc, ok := idx[code]
			if !ok || (acc.Category != CategoryAsset && acc.Category != CategoryLiability) {
				continue
			}
			od, _ := ParseAmount(ln.OriginalDebit)
			oc, _ := ParseAmount(ln.OriginalCredit)
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			k := balKey{code, cur}
			if fcBal[k] == nil {
				fcBal[k] = big.NewInt(0)
				bookBal[k] = big.NewInt(0)
			}
			switch acc.Category {
			case CategoryAsset:
				fcBal[k].Add(fcBal[k], od)
				fcBal[k].Sub(fcBal[k], oc)
				bookBal[k].Add(bookBal[k], d)
				bookBal[k].Sub(bookBal[k], c)
			case CategoryLiability:
				fcBal[k].Add(fcBal[k], oc)
				fcBal[k].Sub(fcBal[k], od)
				bookBal[k].Add(bookBal[k], c)
				bookBal[k].Sub(bookBal[k], d)
			}
		}
	}

	var out []FCBalance
	for k, fc := range fcBal {
		if fc.Sign() == 0 {
			continue
		}
		bk := bookBal[k]
		name := k.code
		if acc, ok := idx[k.code]; ok {
			name = acc.Name
		}
		row := FCBalance{
			AccountCode:    k.code,
			AccountName:    name,
			Currency:       k.cur,
			ForeignBalance: formatCents(fc),
			BookBalance:    formatCents(bk),
		}
		if fc.Sign() != 0 {
			rat := new(big.Rat).SetFrac(bk, fc)
			row.ImpliedRate = rat.FloatString(6)
		}
		out = append(out, row)
	}
	return out
}

// BuildRevaluationReport applies period-end rates to FC balances.
func BuildRevaluationReport(settings CurrencySettings, balances []FCBalance, rates PeriodFxRates) RevaluationReport {
	rateMap := map[string]string{}
	for _, r := range rates.Rates {
		rateMap[strings.ToUpper(r.Currency)] = strings.TrimSpace(r.Rate)
	}
	var lines []RevaluationLine
	total := big.NewInt(0)
	for _, b := range balances {
		fc, _ := ParseAmount(b.ForeignBalance)
		book, _ := ParseAmount(b.BookBalance)
		if fc.Sign() == 0 {
			continue
		}
		rateStr, ok := rateMap[b.Currency]
		if !ok || rateStr == "" {
			continue
		}
		reval := mulFCByRate(fc, rateStr)
		gl := new(big.Int).Sub(reval, book)
		lines = append(lines, RevaluationLine{
			AccountCode:     b.AccountCode,
			AccountName:     b.AccountName,
			Currency:        b.Currency,
			ForeignBalance:  b.ForeignBalance,
			BookBalance:     b.BookBalance,
			ClosingRate:     rateStr,
			RevaluedBalance: formatCents(reval),
			GainLoss:        formatCents(gl),
		})
		total.Add(total, gl)
	}
	return RevaluationReport{
		Period:          rates.Period,
		BaseCurrency:    settings.BaseCurrency,
		Lines:           lines,
		TotalGainLoss:   formatCents(total),
		GainLossAccount: settings.GainLossAccount,
	}
}

// mulFCByRate: base_cents = (fc_cents/100) * rate * 100 = fc_cents * rate (rate decimal).
func mulFCByRate(fcCents *big.Int, rateStr string) *big.Int {
	rat, ok := new(big.Rat).SetString(strings.TrimSpace(rateStr))
	if !ok || fcCents.Sign() == 0 {
		return big.NewInt(0)
	}
	fc := new(big.Rat).SetFrac(fcCents, big.NewInt(100))
	fc.Mul(fc, rat)
	fc.Mul(fc, big.NewRat(100, 1))
	out, _ := fc.Float64()
	return big.NewInt(int64(out + 0.5))
}
