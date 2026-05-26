package accounting

import (
	"strings"
	"time"
)

// DefaultChart returns a minimal Chinese chart of accounts for small ledgers.
func DefaultChart(ledgerID string) ChartOfAccounts {
	return ChartOfAccounts{
		LedgerID: ledgerID,
		Accounts: []Account{
			{Code: "1001", Name: "库存现金", Category: CategoryAsset, Active: true},
			{Code: "1002", Name: "银行存款", Category: CategoryAsset, Active: true},
			{Code: "1122", Name: "应收账款", Category: CategoryAsset, Active: true},
			{Code: "2202", Name: "应付账款", Category: CategoryLiability, Active: true},
			{Code: "4001", Name: "实收资本", Category: CategoryEquity, Active: true},
			{Code: "6001", Name: "主营业务收入", Category: CategoryRevenue, Active: true},
			{Code: "6401", Name: "主营业务成本", Category: CategoryExpense, Active: true},
			{Code: "6601", Name: "销售费用", Category: CategoryExpense, Active: true},
			{Code: "6602", Name: "管理费用", Category: CategoryExpense, Active: true},
		},
		UpdatedAt: time.Now().UTC(),
	}
}

// ValidateChart checks account codes and categories.
func ValidateChart(c *ChartOfAccounts) error {
	if c == nil || len(c.Accounts) == 0 {
		return ErrInvalidAccount
	}
	seen := map[string]bool{}
	for _, a := range c.Accounts {
		code := strings.TrimSpace(a.Code)
		if code == "" || seen[code] {
			return ErrInvalidAccount
		}
		if strings.TrimSpace(a.Name) == "" {
			return ErrInvalidAccount
		}
		switch a.Category {
		case CategoryAsset, CategoryLiability, CategoryEquity, CategoryRevenue, CategoryExpense:
		default:
			return ErrInvalidAccount
		}
		seen[code] = true
	}
	return nil
}

// AccountIndex builds code -> account map (active only).
func AccountIndex(c ChartOfAccounts) map[string]Account {
	out := make(map[string]Account, len(c.Accounts))
	for _, a := range c.Accounts {
		if a.Active {
			out[a.Code] = a
		}
	}
	return out
}
