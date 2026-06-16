package accounting

import (
	"fmt"
	"math/big"
	"strings"
)

// ParseAmount converts decimal string to cents (×100).
func ParseAmount(s string) (*big.Int, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return big.NewInt(0), nil
	}
	rat, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, fmt.Errorf("invalid amount %q", s)
	}
	rat.Mul(rat, big.NewRat(100, 1))
	f := new(big.Float).SetRat(rat)
	i, _ := f.Int(nil)
	return i, nil
}

// ValidateJournal checks lines balance and account codes exist.
func ValidateJournal(j *JournalEntry, chart ChartOfAccounts) error {
	if j == nil || len(j.Lines) < 2 {
		return ErrInvalidJournal
	}
	if _, err := PeriodFromDate(j.Date); err != nil {
		return err
	}
	idx := AccountIndex(chart)
	var debitSum, creditSum big.Int
	for _, ln := range j.Lines {
		code := strings.TrimSpace(ln.AccountCode)
		if code == "" {
			return ErrInvalidJournal
		}
		if _, ok := idx[code]; !ok {
			return fmt.Errorf("%w: %s", ErrAccountNotFound, code)
		}
		d, err := ParseAmount(ln.Debit)
		if err != nil {
			return err
		}
		c, err := ParseAmount(ln.Credit)
		if err != nil {
			return err
		}
		if d.Sign() > 0 && c.Sign() > 0 {
			return ErrInvalidJournal
		}
		if d.Sign() < 0 || c.Sign() < 0 {
			return ErrInvalidJournal
		}
		debitSum.Add(&debitSum, d)
		creditSum.Add(&creditSum, c)
	}
	if debitSum.Cmp(&creditSum) != 0 {
		return ErrUnbalanced
	}
	if debitSum.Sign() == 0 {
		return ErrUnbalanced
	}
	return nil
}

// LineTotals returns debit/credit sums as decimal strings.
func LineTotals(lines []JournalLine) (debit, credit string, err error) {
	var dSum, cSum big.Int
	for _, ln := range lines {
		d, err := ParseAmount(ln.Debit)
		if err != nil {
			return "", "", err
		}
		c, err := ParseAmount(ln.Credit)
		if err != nil {
			return "", "", err
		}
		dSum.Add(&dSum, d)
		cSum.Add(&cSum, c)
	}
	return formatCents(&dSum), formatCents(&cSum), nil
}

func formatCents(c *big.Int) string {
	if c.Sign() == 0 {
		return "0"
	}
	r := new(big.Rat).SetFrac(c, big.NewInt(100))
	return r.FloatString(2)
}
