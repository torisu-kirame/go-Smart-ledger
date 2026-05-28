package accounting

import (
	"fmt"
	"math/big"
	"strings"
)

// ValidateBudget checks period budget payload.
func ValidateBudget(b *PeriodBudget) error {
	if b == nil {
		return ErrInvalidBudget
	}
	if _, err := NormalizePeriod(b.Period); err != nil {
		return ErrInvalidBudget
	}
	seen := map[string]bool{}
	for _, ln := range b.Lines {
		key := string(ln.Scope) + ":" + strings.TrimSpace(ln.ScopeKey)
		if strings.TrimSpace(ln.ScopeKey) == "" || seen[key] {
			return ErrInvalidBudget
		}
		if ln.Scope != BudgetScopeAccount && ln.Scope != BudgetScopeProject {
			return ErrInvalidBudget
		}
		amt, err := ParseAmount(ln.Amount)
		if err != nil || amt.Sign() < 0 {
			return ErrInvalidBudget
		}
		seen[key] = true
	}
	return nil
}

// BuildBudgetAnalysis compares budget lines to journal actuals in the period.
func BuildBudgetAnalysis(chart ChartOfAccounts, budget PeriodBudget, journals []JournalEntry) BudgetAnalysis {
	idx := AccountIndex(chart)
	out := BudgetAnalysis{Period: budget.Period, Lines: make([]BudgetExecutionLine, 0, len(budget.Lines))}
	for _, bl := range budget.Lines {
		budgetAmt, _ := ParseAmount(bl.Amount)
		actual := big.NewInt(0)
		label := bl.ScopeKey
		switch bl.Scope {
		case BudgetScopeAccount:
			if acc, ok := idx[bl.ScopeKey]; ok {
				label = acc.Code + " " + acc.Name
			}
			actual = actualForAccount(bl.ScopeKey, idx, journals, budget.Period)
		case BudgetScopeProject:
			label = "项目 " + bl.ScopeKey
			actual = actualForProject(bl.ScopeKey, idx, journals, budget.Period)
		}
		variance := new(big.Int).Sub(actual, budgetAmt)
		line := BudgetExecutionLine{
			Scope:      bl.Scope,
			ScopeKey:   bl.ScopeKey,
			ScopeLabel: label,
			Budget:     formatCents(budgetAmt),
			Actual:     formatCents(actual),
			Variance:   formatCents(variance),
			OverBudget: actual.Cmp(budgetAmt) > 0,
		}
		if budgetAmt.Sign() > 0 {
			pct := new(big.Rat).SetFrac(new(big.Int).Mul(actual, big.NewInt(100)), budgetAmt)
			line.UtilizationPct = pct.FloatString(1) + "%"
		}
		out.Lines = append(out.Lines, line)
		if line.OverBudget {
			out.Alerts = append(out.Alerts, line)
		}
	}
	return out
}

func actualForAccount(code string, idx map[string]Account, journals []JournalEntry, period string) *big.Int {
	acc, ok := idx[code]
	if !ok {
		return big.NewInt(0)
	}
	sum := big.NewInt(0)
	for _, j := range journals {
		if j.Period != period {
			continue
		}
		for _, ln := range j.Lines {
			if strings.TrimSpace(ln.AccountCode) != code {
				continue
			}
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			switch acc.Category {
			case CategoryRevenue:
				sum.Add(sum, c)
				sum.Sub(sum, d)
			default:
				// expense & others: debit increases spend
				sum.Add(sum, d)
				sum.Sub(sum, c)
			}
		}
	}
	if sum.Sign() < 0 {
		return big.NewInt(0)
	}
	return sum
}

func actualForProject(project string, idx map[string]Account, journals []JournalEntry, period string) *big.Int {
	project = strings.TrimSpace(project)
	sum := big.NewInt(0)
	for _, j := range journals {
		if j.Period != period {
			continue
		}
		for _, ln := range j.Lines {
			if strings.TrimSpace(ln.Project) != project {
				continue
			}
			acc, ok := idx[ln.AccountCode]
			if !ok {
				continue
			}
			if acc.Category != CategoryExpense && acc.Category != CategoryRevenue {
				continue
			}
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			if acc.Category == CategoryRevenue {
				sum.Add(sum, c)
				sum.Sub(sum, d)
			} else {
				sum.Add(sum, d)
				sum.Sub(sum, c)
			}
		}
	}
	if sum.Sign() < 0 {
		return big.NewInt(0)
	}
	return sum
}

// EmptyBudget returns zero-line budget for a period.
func EmptyBudget(ledgerID, period string) PeriodBudget {
	return PeriodBudget{LedgerID: ledgerID, Period: period, Lines: []BudgetLine{}}
}

// BudgetLineKey for dedup.
func BudgetLineKey(scope BudgetScope, key string) string {
	return fmt.Sprintf("%s:%s", scope, strings.TrimSpace(key))
}
