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
	ErrInvalidCurrency  = errors.New("invalid currency settings")
	ErrInvalidFxRates   = errors.New("invalid fx rates")
	ErrInvalidTax       = errors.New("invalid tax template")
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
	Currency     string `json:"currency,omitempty"`     // 原币 ISO（F46）
	FxRate       string `json:"fxRate,omitempty"`       // 1 原币 = ? 本位币
	OriginalDebit  string `json:"originalDebit,omitempty"`
	OriginalCredit string `json:"originalCredit,omitempty"`
	TaxCategory  string `json:"taxCategory,omitempty"` // taxable|exempt|zero|none（F47）
	TaxRate      string `json:"taxRate,omitempty"`
	TaxAmount    string `json:"taxAmount,omitempty"`
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

// CurrencySettings per ledger (F46).
type CurrencySettings struct {
	LedgerID          string   `json:"ledgerId"`
	BaseCurrency      string   `json:"baseCurrency"`
	Currencies        []string `json:"currencies"`
	GainLossAccount   string   `json:"gainLossAccount"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// FxRateRow is closing rate: 1 unit of Currency = Rate units of base.
type FxRateRow struct {
	Currency string `json:"currency"`
	Rate     string `json:"rate"`
}

// PeriodFxRates stores month-end rates (F46).
type PeriodFxRates struct {
	LedgerID     string      `json:"ledgerId"`
	Period       string      `json:"period"`
	BaseCurrency string      `json:"baseCurrency"`
	Rates        []FxRateRow `json:"rates"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// FCBalance is foreign-currency balance on a monetary account.
type FCBalance struct {
	AccountCode    string `json:"accountCode"`
	AccountName    string `json:"accountName"`
	Currency       string `json:"currency"`
	ForeignBalance string `json:"foreignBalance"`
	BookBalance    string `json:"bookBalance"`
	ImpliedRate    string `json:"impliedRate,omitempty"`
}

// RevaluationLine is one FC position revalued at period end.
type RevaluationLine struct {
	AccountCode     string `json:"accountCode"`
	AccountName     string `json:"accountName"`
	Currency        string `json:"currency"`
	ForeignBalance  string `json:"foreignBalance"`
	BookBalance     string `json:"bookBalance"`
	ClosingRate     string `json:"closingRate"`
	RevaluedBalance  string `json:"revaluedBalance"`
	GainLoss        string `json:"gainLoss"`
}

// RevaluationReport summarizes FX revaluation (F46).
type RevaluationReport struct {
	Period          string            `json:"period"`
	BaseCurrency    string            `json:"baseCurrency"`
	Lines           []RevaluationLine `json:"lines"`
	TotalGainLoss   string            `json:"totalGainLoss"`
	GainLossAccount string            `json:"gainLossAccount"`
}

// TaxMode for VAT templates (F47).
type TaxMode string

const (
	TaxModeNone    TaxMode = "none"
	TaxModeGeneral TaxMode = "general"
	TaxModeSimple  TaxMode = "simple"
)

// TaxTemplate is ledger tax policy.
type TaxTemplate struct {
	LedgerID            string    `json:"ledgerId"`
	Mode                TaxMode   `json:"mode"`
	Region              string    `json:"region,omitempty"`
	DefaultOutputRate   string    `json:"defaultOutputRate,omitempty"`
	DefaultInputRate    string    `json:"defaultInputRate,omitempty"`
	SimpleLevyRate      string    `json:"simpleLevyRate,omitempty"`
	OutputTaxAccount    string    `json:"outputTaxAccount,omitempty"`
	InputTaxAccount     string    `json:"inputTaxAccount,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// BuiltinTaxPreset is a selectable template id + defaults.
type BuiltinTaxPreset struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Mode                TaxMode `json:"mode"`
	DefaultOutputRate   string  `json:"defaultOutputRate,omitempty"`
	DefaultInputRate    string  `json:"defaultInputRate,omitempty"`
	SimpleLevyRate      string  `json:"simpleLevyRate,omitempty"`
}

// TaxReportLine is one taxable journal line in a period.
type TaxReportLine struct {
	JournalID    string `json:"journalId"`
	Date         string `json:"date"`
	Description  string `json:"description,omitempty"`
	AccountCode  string `json:"accountCode"`
	AccountName  string `json:"accountName"`
	BaseAmount   string `json:"baseAmount"`
	TaxCategory  string `json:"taxCategory"`
	TaxRate      string `json:"taxRate"`
	TaxAmount    string `json:"taxAmount"`
	Kind         string `json:"kind"` // output | input | levy
}

// TaxReport is period VAT / levy summary (F47).
type TaxReport struct {
	Period          string          `json:"period"`
	Mode            TaxMode         `json:"mode"`
	OutputTaxTotal  string          `json:"outputTaxTotal"`
	InputTaxTotal   string          `json:"inputTaxTotal"`
	NetPayable      string          `json:"netPayable"`
	Lines           []TaxReportLine `json:"lines"`
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
