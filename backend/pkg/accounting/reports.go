package accounting

import (
	"math/big"
	"sort"
)

// AccountBalance is ending balance for one account.
type AccountBalance struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	Category AccountCategory `json:"category"`
	Debit    string          `json:"debit"`
	Credit   string          `json:"credit"`
	Balance  string          `json:"balance"` // signed: + debit nature, - credit nature for display
}

// TrialBalanceReport lists all accounts with activity.
type TrialBalanceReport struct {
	Period   string           `json:"period,omitempty"`
	Accounts []AccountBalance `json:"accounts"`
	Balanced bool             `json:"balanced"`
}

// FinancialReports bundles the three statements (simplified).
type FinancialReports struct {
	Period        string           `json:"period,omitempty"`
	TrialBalance  TrialBalanceReport `json:"trialBalance"`
	BalanceSheet  StatementSection `json:"balanceSheet"`
	IncomeStatement StatementSection `json:"incomeStatement"`
	CashFlow      CashFlowReport   `json:"cashFlow"`
}

// StatementSection is a grouped list of lines.
type StatementSection struct {
	Lines []StatementLine `json:"lines"`
	Total string          `json:"total"`
}

// StatementLine one row on a report.
type StatementLine struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Amount  string `json:"amount"`
}

// CashFlowReport simplified direct method.
type CashFlowReport struct {
	Operating  string `json:"operating"`
	Investing  string `json:"investing"`
	Financing  string `json:"financing"`
	NetChange  string `json:"netChange"`
}

// BuildTrialBalance aggregates journals into account balances.
func BuildTrialBalance(chart ChartOfAccounts, journals []JournalEntry, throughPeriod string) TrialBalanceReport {
	idx := AccountIndex(chart)
	bal := map[string]*big.Int{}
	for code := range idx {
		bal[code] = big.NewInt(0)
	}
	for _, j := range journals {
		if throughPeriod != "" && j.Period > throughPeriod {
			continue
		}
		for _, ln := range j.Lines {
			code := ln.AccountCode
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			if bal[code] == nil {
				bal[code] = big.NewInt(0)
			}
			bal[code].Add(bal[code], d)
			bal[code].Sub(bal[code], c)
		}
	}
	var out []AccountBalance
	var totalDebit, totalCredit big.Int
	for code, acc := range idx {
		net := bal[code]
		if net == nil || net.Sign() == 0 {
			continue
		}
		ab := AccountBalance{Code: code, Name: acc.Name, Category: acc.Category}
		if net.Sign() > 0 {
			ab.Debit = formatCents(net)
			totalDebit.Add(&totalDebit, net)
		} else {
			neg := new(big.Int).Neg(net)
			ab.Credit = formatCents(neg)
			totalCredit.Add(&totalCredit, neg)
		}
		ab.Balance = formatCents(net)
		out = append(out, ab)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return TrialBalanceReport{
		Period:   throughPeriod,
		Accounts: out,
		Balanced: totalDebit.Cmp(&totalCredit) == 0,
	}
}

// BuildFinancialReports produces balance sheet, P&L, and simplified cash flow.
func BuildFinancialReports(chart ChartOfAccounts, journals []JournalEntry, period string) FinancialReports {
	tb := BuildTrialBalance(chart, journals, period)
	var assets, liabilities, equity, revenue, expense big.Int
	for _, ab := range tb.Accounts {
		net, _ := ParseAmount(ab.Balance)
		switch ab.Category {
		case CategoryAsset:
			assets.Add(&assets, net)
		case CategoryLiability:
			liabilities.Sub(&liabilities, net)
		case CategoryEquity:
			equity.Sub(&equity, net)
		case CategoryRevenue:
			revenue.Sub(&revenue, net)
		case CategoryExpense:
			expense.Add(&expense, net)
		}
	}
	bs := StatementSection{Lines: sectionLines(tb.Accounts, CategoryAsset, CategoryLiability, CategoryEquity)}
	bs.Total = formatCents(new(big.Int).Add(&assets, new(big.Int).Add(&liabilities, &equity)))

	pl := StatementSection{Lines: sectionLines(tb.Accounts, CategoryRevenue, CategoryExpense)}
	netIncome := new(big.Int).Sub(&revenue, &expense)
	pl.Total = formatCents(netIncome)

	cashCodes := map[string]bool{"1001": true, "1002": true}
	var cashChange big.Int
	for _, j := range journals {
		if period != "" && j.Period != period {
			continue
		}
		for _, ln := range j.Lines {
			if !cashCodes[ln.AccountCode] {
				continue
			}
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			cashChange.Add(&cashChange, d)
			cashChange.Sub(&cashChange, c)
		}
	}

	return FinancialReports{
		Period:          period,
		TrialBalance:    tb,
		BalanceSheet:    bs,
		IncomeStatement: pl,
		CashFlow: CashFlowReport{
			Operating: formatCents(&cashChange),
			NetChange: formatCents(&cashChange),
		},
	}
}

func sectionLines(accounts []AccountBalance, cats ...AccountCategory) []StatementLine {
	cset := map[AccountCategory]bool{}
	for _, c := range cats {
		cset[c] = true
	}
	var lines []StatementLine
	for _, ab := range accounts {
		if !cset[ab.Category] {
			continue
		}
		amt, _ := ParseAmount(ab.Balance)
		if amt.Sign() == 0 {
			continue
		}
		lines = append(lines, StatementLine{Code: ab.Code, Name: ab.Name, Amount: formatCents(amt)})
	}
	return lines
}
