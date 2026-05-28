package accounting

import (
	"math/big"
	"strings"
	"time"
)

// BuiltinTaxPresets returns selectable CN VAT templates.
func BuiltinTaxPresets() []BuiltinTaxPreset {
	return []BuiltinTaxPreset{
		{
			ID: "cn_vat_general", Name: "增值税 · 一般纳税人（13%）",
			Mode: TaxModeGeneral, DefaultOutputRate: "0.13", DefaultInputRate: "0.13",
		},
		{
			ID: "cn_vat_general_9", Name: "增值税 · 一般纳税人（9% 销项）",
			Mode: TaxModeGeneral, DefaultOutputRate: "0.09", DefaultInputRate: "0.13",
		},
		{
			ID: "cn_vat_simple", Name: "增值税 · 简易计税（3%）",
			Mode: TaxModeSimple, SimpleLevyRate: "0.03",
		},
		{
			ID: "cn_vat_exempt", Name: "免税 / 不征税",
			Mode: TaxModeNone,
		},
	}
}

// DefaultTaxTemplate for a ledger.
func DefaultTaxTemplate(ledgerID string) TaxTemplate {
	return TaxTemplate{
		LedgerID:          ledgerID,
		Mode:              TaxModeGeneral,
		Region:            "CN",
		DefaultOutputRate: "0.13",
		DefaultInputRate:  "0.13",
		OutputTaxAccount:  "2221",
		InputTaxAccount:   "2221",
		UpdatedAt:         time.Now().UTC(),
	}
}

// ValidateTaxTemplate checks tax settings.
func ValidateTaxTemplate(t *TaxTemplate) error {
	if t == nil {
		return ErrInvalidTax
	}
	switch t.Mode {
	case TaxModeNone, TaxModeGeneral, TaxModeSimple:
	default:
		return ErrInvalidTax
	}
	if t.Mode == TaxModeGeneral {
		if _, err := parseRateFraction(t.DefaultOutputRate); err != nil {
			return ErrInvalidTax
		}
		if _, err := parseRateFraction(t.DefaultInputRate); err != nil {
			return ErrInvalidTax
		}
	}
	if t.Mode == TaxModeSimple {
		if _, err := parseRateFraction(t.SimpleLevyRate); err != nil {
			return ErrInvalidTax
		}
	}
	if strings.TrimSpace(t.OutputTaxAccount) == "" {
		t.OutputTaxAccount = "2221"
	}
	if strings.TrimSpace(t.InputTaxAccount) == "" {
		t.InputTaxAccount = "2221"
	}
	return nil
}

// ApplyTaxPreset copies preset fields onto template.
func ApplyTaxPreset(t *TaxTemplate, presetID string) bool {
	for _, p := range BuiltinTaxPresets() {
		if p.ID != presetID {
			continue
		}
		t.Mode = p.Mode
		t.DefaultOutputRate = p.DefaultOutputRate
		t.DefaultInputRate = p.DefaultInputRate
		t.SimpleLevyRate = p.SimpleLevyRate
		return true
	}
	return false
}

// BuildTaxReport aggregates VAT from journals in a period.
func BuildTaxReport(chart ChartOfAccounts, tmpl TaxTemplate, journals []JournalEntry, period string) TaxReport {
	idx := AccountIndex(chart)
	rep := TaxReport{
		Period:         period,
		Mode:           tmpl.Mode,
		OutputTaxTotal: "0",
		InputTaxTotal:  "0",
		NetPayable:     "0",
		Lines:          nil,
	}
	if tmpl.Mode == TaxModeNone {
		return rep
	}
	outSum := big.NewInt(0)
	inSum := big.NewInt(0)

	for _, j := range journals {
		if period != "" && j.Period != period {
			continue
		}
		for _, ln := range j.Lines {
			acc, ok := idx[ln.AccountCode]
			if !ok {
				continue
			}
			cat := strings.ToLower(strings.TrimSpace(ln.TaxCategory))
			if cat == "" || cat == "none" {
				if tmpl.Mode == TaxModeSimple && acc.Category == CategoryRevenue {
					cat = "taxable"
				} else if tmpl.Mode == TaxModeGeneral && acc.Category == CategoryRevenue {
					cat = "taxable"
				} else if tmpl.Mode == TaxModeGeneral && acc.Category == CategoryExpense {
					cat = "taxable"
				} else {
					continue
				}
			}
			if cat == "exempt" || cat == "zero" {
				continue
			}
			base := lineBaseAmount(acc.Category, ln)
			if base.Sign() == 0 {
				continue
			}

			var taxAmt *big.Int
			var kind string
			rateStr := ln.TaxRate
			if rateStr == "" {
				switch tmpl.Mode {
				case TaxModeSimple:
					rateStr = tmpl.SimpleLevyRate
				case TaxModeGeneral:
					if acc.Category == CategoryRevenue {
						rateStr = tmpl.DefaultOutputRate
					} else if acc.Category == CategoryExpense {
						rateStr = tmpl.DefaultInputRate
					}
				}
			}
			if ln.TaxAmount != "" {
				taxAmt, _ = ParseAmount(ln.TaxAmount)
			} else {
				taxAmt = taxFromBase(base, rateStr)
			}
			if taxAmt.Sign() == 0 {
				continue
			}

			switch tmpl.Mode {
			case TaxModeSimple:
				if acc.Category != CategoryRevenue {
					continue
				}
				kind = "levy"
				outSum.Add(outSum, taxAmt)
			case TaxModeGeneral:
				if acc.Category == CategoryRevenue {
					kind = "output"
					outSum.Add(outSum, taxAmt)
				} else if acc.Category == CategoryExpense {
					kind = "input"
					inSum.Add(inSum, taxAmt)
				} else {
					continue
				}
			}

			name := acc.Name
			rep.Lines = append(rep.Lines, TaxReportLine{
				JournalID:   j.ID,
				Date:        j.Date,
				Description: j.Description,
				AccountCode: ln.AccountCode,
				AccountName: name,
				BaseAmount:  formatCents(base),
				TaxCategory: effectiveCategory(cat, tmpl.Mode),
				TaxRate:     rateStr,
				TaxAmount:   formatCents(taxAmt),
				Kind:        kind,
			})
		}
	}
	rep.OutputTaxTotal = formatCents(outSum)
	rep.InputTaxTotal = formatCents(inSum)
	net := new(big.Int).Sub(outSum, inSum)
	rep.NetPayable = formatCents(net)
	return rep
}

func effectiveCategory(cat string, mode TaxMode) string {
	if cat != "" {
		return cat
	}
	if mode == TaxModeSimple {
		return "taxable"
	}
	return "taxable"
}

func lineBaseAmount(cat AccountCategory, ln JournalLine) *big.Int {
	d, _ := ParseAmount(ln.Debit)
	c, _ := ParseAmount(ln.Credit)
	switch cat {
	case CategoryRevenue:
		return new(big.Int).Sub(c, d)
	case CategoryExpense:
		return new(big.Int).Sub(d, c)
	default:
		return big.NewInt(0)
	}
}

func taxFromBase(base *big.Int, rateStr string) *big.Int {
	rat, err := parseRateFraction(rateStr)
	if err != nil || base.Sign() == 0 {
		return big.NewInt(0)
	}
	// tax = base * rate (base in cents)
	out := new(big.Rat).SetFrac(base, big.NewInt(1))
	out.Mul(out, rat)
	f, _ := out.Float64()
	return big.NewInt(int64(f + 0.5))
}

func parseRateFraction(s string) (*big.Rat, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidTax
	}
	if strings.HasPrefix(s, "0.") || strings.HasPrefix(s, ".") {
		r, ok := new(big.Rat).SetString(s)
		if !ok {
			return nil, ErrInvalidTax
		}
		return r, nil
	}
	// percent like 13 -> 0.13
	amt, err := ParseAmount(s)
	if err != nil {
		return nil, err
	}
	if amt.Cmp(big.NewInt(100)) <= 0 {
		return new(big.Rat).SetFrac(amt, big.NewInt(100)), nil
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, ErrInvalidTax
	}
	return r, nil
}
