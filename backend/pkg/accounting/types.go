package accounting

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidAccount   = errors.New("invalid account")
	ErrUnbalanced       = errors.New("journal entry not balanced")
	ErrPeriodClosed     = errors.New("accounting period is closed")
	ErrPeriodNotFound   = errors.New("accounting period not found")
	ErrInvalidPeriod    = errors.New("invalid period")
	ErrInvalidJournal   = errors.New("invalid journal entry")
	ErrAccountNotFound  = errors.New("account not found")
	ErrStmtNotFound     = errors.New("bank statement not found")
	ErrLineNotFound     = errors.New("statement line not found")
	ErrInvalidBudget    = errors.New("invalid budget")
)

// AccountCategory for financial statement grouping.
type AccountCategory string

const (
	CategoryAsset     AccountCategory = "asset"
	CategoryLiability AccountCategory = "liability"
	CategoryEquity    AccountCategory = "equity"
	CategoryRevenue   AccountCategory = "revenue"
	CategoryExpense   AccountCategory = "expense"
)

// Account is one row in the chart of accounts.
type Account struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	Category AccountCategory `json:"category"`
	Active   bool            `json:"active"`
}

// ChartOfAccounts is stored per ledger on chain.
type ChartOfAccounts struct {
	LedgerID  string    `json:"ledgerId"`
	Accounts  []Account `json:"accounts"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// JournalLine is one debit or credit line.
type JournalLine struct {
	AccountCode  string `json:"accountCode"`
	Debit        string `json:"debit,omitempty"`  // decimal string
	Credit       string `json:"credit,omitempty"` // decimal string
	Memo         string `json:"memo,omitempty"`
	Counterparty string `json:"counterparty,omitempty"` // 往来方（F45 账龄）
	Project      string `json:"project,omitempty"`      // 项目（F43 预算）
}

// JournalEntry is a double-entry voucher.
type JournalEntry struct {
	ID          string        `json:"id"`
	LedgerID    string        `json:"ledgerId"`
	Period      string        `json:"period"` // YYYY-MM
	Date        string        `json:"date"`   // YYYY-MM-DD
	Description string        `json:"description,omitempty"`
	Lines       []JournalLine `json:"lines"`
	PostedBy    string        `json:"postedBy"`
	PostedAt    time.Time     `json:"postedAt"`
	EventSeq    uint64        `json:"eventSeq,omitempty"`
}

// PeriodStatus for month-end control.
type PeriodStatus string

const (
	PeriodOpen   PeriodStatus = "open"
	PeriodClosed PeriodStatus = "closed"
)

// PeriodState stored per YYYY-MM.
type PeriodState struct {
	Period    string       `json:"period"`
	Status    PeriodStatus `json:"status"`
	ClosedAt  time.Time    `json:"closedAt,omitempty"`
	ClosedBy  string       `json:"closedBy,omitempty"`
}

// AuxiliaryDims are optional analytic tags on vouchers/attachments (F41).
type AuxiliaryDims struct {
	Department   string `json:"department,omitempty"`   // 部门
	Project      string `json:"project,omitempty"`      // 项目
	Counterparty string `json:"counterparty,omitempty"` // 往来
}

// Attachment links a file (IPFS CID) to an entry event seq.
type Attachment struct {
	ID         string         `json:"id"`
	TableID    string         `json:"tableId,omitempty"`
	EntrySeq   uint64         `json:"entrySeq"`
	Filename   string         `json:"filename"`
	MimeType   string         `json:"mimeType,omitempty"`
	Size       int64          `json:"size,omitempty"`
	CID        string         `json:"cid"`
	Ref        string         `json:"ref,omitempty"`
	Auxiliary  *AuxiliaryDims `json:"auxiliary,omitempty"`
	UploadedBy string         `json:"uploadedBy"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// BankStatementLine is one imported bank row.
type BankStatementLine struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description,omitempty"`
	Amount      string `json:"amount"` // signed: + inflow, - outflow
	Balance     string `json:"balance,omitempty"`
	MatchedSeq  uint64 `json:"matchedSeq,omitempty"`
}

// BudgetScope is account- or project-based budget (F43).
type BudgetScope string

const (
	BudgetScopeAccount BudgetScope = "account"
	BudgetScopeProject BudgetScope = "project"
)

// BudgetLine is one budget row for a period.
type BudgetLine struct {
	ID       string      `json:"id"`
	Scope    BudgetScope `json:"scope"`
	ScopeKey string      `json:"scopeKey"`
	Amount   string      `json:"amount"`
}

// PeriodBudget stored per ledger per YYYY-MM.
type PeriodBudget struct {
	LedgerID  string       `json:"ledgerId"`
	Period    string       `json:"period"`
	Lines     []BudgetLine `json:"lines"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// BudgetExecutionLine compares budget vs actual for one scope.
type BudgetExecutionLine struct {
	Scope          BudgetScope `json:"scope"`
	ScopeKey       string      `json:"scopeKey"`
	ScopeLabel     string      `json:"scopeLabel,omitempty"`
	Budget         string      `json:"budget"`
	Actual         string      `json:"actual"`
	Variance       string      `json:"variance"`
	UtilizationPct string      `json:"utilizationPct,omitempty"`
	OverBudget     bool        `json:"overBudget"`
}

// BudgetAnalysis is execution report for a period (F43).
type BudgetAnalysis struct {
	Period  string                `json:"period"`
	Lines   []BudgetExecutionLine `json:"lines"`
	Alerts  []BudgetExecutionLine `json:"alerts,omitempty"`
}

// AgingKind distinguishes receivable vs payable aging (F45).
type AgingKind string

const (
	AgingReceivable AgingKind = "receivable"
	AgingPayable    AgingKind = "payable"
)

// AgingCounterpartySummary buckets open balances by counterparty.
type AgingCounterpartySummary struct {
	Counterparty string    `json:"counterparty"`
	Kind         AgingKind `json:"kind"`
	AccountCode  string    `json:"accountCode"`
	AccountName  string    `json:"accountName"`
	Total        string    `json:"total"`
	Current      string    `json:"current"`
	Days31_60    string    `json:"days31_60"`
	Days61_90    string    `json:"days61_90"`
	Over90       string    `json:"over90"`
}

// AgingOpenItem is one unmatched open amount.
type AgingOpenItem struct {
	Counterparty string    `json:"counterparty"`
	Kind         AgingKind `json:"kind"`
	AccountCode  string    `json:"accountCode"`
	Date         string    `json:"date"`
	Amount       string    `json:"amount"`
	Days         int       `json:"days"`
	Bucket       string    `json:"bucket"`
	JournalID    string    `json:"journalId,omitempty"`
	Description  string    `json:"description,omitempty"`
}

// AgingReport is AR/AP aging as of a date (F45).
type AgingReport struct {
	AsOf                string                     `json:"asOf"`
	ReceivableAccounts  []string                   `json:"receivableAccounts"`
	PayableAccounts     []string                   `json:"payableAccounts"`
	Summaries           []AgingCounterpartySummary `json:"summaries"`
	Items               []AgingOpenItem            `json:"items,omitempty"`
}

// BankStatement groups imported lines.
type BankStatement struct {
	ID        string              `json:"id"`
	LedgerID  string              `json:"ledgerId"`
	AccountCode string            `json:"accountCode,omitempty"`
	Filename  string              `json:"filename,omitempty"`
	Lines     []BankStatementLine `json:"lines"`
	ImportedBy string             `json:"importedBy"`
	ImportedAt time.Time           `json:"importedAt"`
}

// NormalizePeriod validates YYYY-MM.
func NormalizePeriod(p string) (string, error) {
	p = strings.TrimSpace(p)
	if len(p) != 7 || p[4] != '-' {
		return "", ErrInvalidPeriod
	}
	if _, err := time.Parse("2006-01", p); err != nil {
		return "", ErrInvalidPeriod
	}
	return p, nil
}

// PeriodFromDate extracts YYYY-MM from YYYY-MM-DD.
func PeriodFromDate(date string) (string, error) {
	date = strings.TrimSpace(date)
	if len(date) < 7 {
		return "", fmt.Errorf("%w: bad date", ErrInvalidJournal)
	}
	return NormalizePeriod(date[:7])
}
