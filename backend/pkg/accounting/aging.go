package accounting

import (
	"math/big"
	"sort"
	"strings"
	"time"
)

const agingUnspecified = "未指定往来方"

// DefaultReceivableAccounts for AR aging.
func DefaultReceivableAccounts() []string { return []string{"1122"} }

// DefaultPayableAccounts for AP aging.
func DefaultPayableAccounts() []string { return []string{"2202"} }

// BuildAgingReport computes AR/AP aging from posted journals.
func BuildAgingReport(chart ChartOfAccounts, journals []JournalEntry, asOf string, recvCodes, payCodes []string) (AgingReport, error) {
	asOf = strings.TrimSpace(asOf)
	if asOf == "" {
		asOf = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", asOf); err != nil {
		return AgingReport{}, ErrInvalidJournal
	}
	if len(recvCodes) == 0 {
		recvCodes = DefaultReceivableAccounts()
	}
	if len(payCodes) == 0 {
		payCodes = DefaultPayableAccounts()
	}
	idx := AccountIndex(chart)
	asOfTime, _ := time.Parse("2006-01-02", asOf)

	var items []AgingOpenItem
	for _, code := range recvCodes {
		items = append(items, openItemsForAccount(journals, idx, code, AgingReceivable, asOfTime)...)
	}
	for _, code := range payCodes {
		items = append(items, openItemsForAccount(journals, idx, code, AgingPayable, asOfTime)...)
	}

	summaries := summarizeAging(items, idx)
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Kind != summaries[j].Kind {
			return summaries[i].Kind < summaries[j].Kind
		}
		return summaries[i].Counterparty < summaries[j].Counterparty
	})
	sort.Slice(items, func(i, j int) bool {
		if items[i].Counterparty != items[j].Counterparty {
			return items[i].Counterparty < items[j].Counterparty
		}
		return items[i].Date < items[j].Date
	})

	return AgingReport{
		AsOf:               asOf,
		ReceivableAccounts: recvCodes,
		PayableAccounts:    payCodes,
		Summaries:          summaries,
		Items:              items,
	}, nil
}

type fifoItem struct {
	date   string
	amount *big.Int
	jid    string
	desc   string
}

func openItemsForAccount(journals []JournalEntry, idx map[string]Account, code string, kind AgingKind, asOf time.Time) []AgingOpenItem {
	acc, ok := idx[code]
	if !ok {
		return nil
	}
	_ = acc
	type key struct {
		cp string
	}
	queues := map[key][]fifoItem{}

	for _, j := range journals {
		if j.Date > asOf.Format("2006-01-02") {
			continue
		}
		for _, ln := range j.Lines {
			if strings.TrimSpace(ln.AccountCode) != code {
				continue
			}
			cp := strings.TrimSpace(ln.Counterparty)
			if cp == "" {
				cp = strings.TrimSpace(ln.Memo)
			}
			if cp == "" {
				cp = agingUnspecified
			}
			d, _ := ParseAmount(ln.Debit)
			c, _ := ParseAmount(ln.Credit)
			k := key{cp: cp}
			q := queues[k]

			switch kind {
			case AgingReceivable:
				if d.Sign() > 0 {
					q = append(q, fifoItem{date: j.Date, amount: new(big.Int).Set(d), jid: j.ID, desc: j.Description})
				}
				if c.Sign() > 0 {
					q = applyPayment(q, c)
				}
			case AgingPayable:
				if c.Sign() > 0 {
					q = append(q, fifoItem{date: j.Date, amount: new(big.Int).Set(c), jid: j.ID, desc: j.Description})
				}
				if d.Sign() > 0 {
					q = applyPayment(q, d)
				}
			}
			queues[k] = q
		}
	}

	var out []AgingOpenItem
	for k, q := range queues {
		for _, it := range q {
			if it.amount.Sign() <= 0 {
				continue
			}
			days, bucket := agingBucket(it.date, asOf)
			out = append(out, AgingOpenItem{
				Counterparty: k.cp,
				Kind:         kind,
				AccountCode:  code,
				Date:         it.date,
				Amount:       formatCents(it.amount),
				Days:         days,
				Bucket:       bucket,
				JournalID:    it.jid,
				Description:  it.desc,
			})
		}
	}
	return out
}

func applyPayment(q []fifoItem, pay *big.Int) []fifoItem {
	remaining := new(big.Int).Set(pay)
	for i := 0; i < len(q) && remaining.Sign() > 0; i++ {
		if q[i].amount.Sign() <= 0 {
			continue
		}
		if q[i].amount.Cmp(remaining) <= 0 {
			remaining.Sub(remaining, q[i].amount)
			q[i].amount = big.NewInt(0)
		} else {
			q[i].amount.Sub(q[i].amount, remaining)
			remaining = big.NewInt(0)
		}
	}
	// compact
	var next []fifoItem
	for _, it := range q {
		if it.amount.Sign() > 0 {
			next = append(next, it)
		}
	}
	return next
}

func agingBucket(itemDate string, asOf time.Time) (days int, bucket string) {
	d, err := time.Parse("2006-01-02", itemDate)
	if err != nil {
		return 0, "0-30"
	}
	days = int(asOf.Sub(d).Hours() / 24)
	if days < 0 {
		days = 0
	}
	switch {
	case days <= 30:
		return days, "0-30"
	case days <= 60:
		return days, "31-60"
	case days <= 90:
		return days, "61-90"
	default:
		return days, "90+"
	}
}

func summarizeAging(items []AgingOpenItem, idx map[string]Account) []AgingCounterpartySummary {
	type aggKey struct {
		cp, kind, code string
	}
	agg := make(map[aggKey]*AgingCounterpartySummary)

	add := func(k aggKey, bucket string, amt *big.Int) {
		s := agg[k]
		if s == nil {
			name := k.code
			if acc, ok := idx[k.code]; ok {
				name = acc.Name
			}
			s = &AgingCounterpartySummary{
				Counterparty: k.cp,
				Kind:         AgingKind(k.kind),
				AccountCode:  k.code,
				AccountName:  name,
			}
			agg[k] = s
		}
		switch bucket {
		case "0-30":
			addCents(&s.Current, amt)
		case "31-60":
			addCents(&s.Days31_60, amt)
		case "61-90":
			addCents(&s.Days61_90, amt)
		default:
			addCents(&s.Over90, amt)
		}
		addCents(&s.Total, amt)
	}

	for _, it := range items {
		amt, _ := ParseAmount(it.Amount)
		k := aggKey{cp: it.Counterparty, kind: string(it.Kind), code: it.AccountCode}
		add(k, it.Bucket, amt)
	}
	out := make([]AgingCounterpartySummary, 0, len(agg))
	for _, s := range agg {
		out = append(out, *s)
	}
	return out
}

func addCents(field *string, amt *big.Int) {
	cur, _ := ParseAmount(*field)
	cur.Add(cur, amt)
	*field = formatCents(cur)
}
