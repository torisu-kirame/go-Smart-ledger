package ledgersvc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/storage"
)

// GetChart returns COA or default if unset.
func (s *Service) GetChart(ctx context.Context, ledgerID, userID string) (*accounting.ChartOfAccounts, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	var chart accounting.ChartOfAccounts
	if err := s.loadJSON(ctx, domain.LedgerCOAKey(ledgerID), &chart); err != nil || len(chart.Accounts) == 0 {
		d := accounting.DefaultChart(ledgerID)
		return &d, nil
	}
	chart.LedgerID = ledgerID
	return &chart, nil
}

// PutChart replaces chart of accounts.
func (s *Service) PutChart(ctx context.Context, ledgerID, userID string, chart accounting.ChartOfAccounts) (*accounting.ChartOfAccounts, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	chart.LedgerID = ledgerID
	chart.UpdatedAt = time.Now().UTC()
	if err := accounting.ValidateChart(&chart); err != nil {
		return nil, err
	}
	if err := s.putJSON(ctx, domain.LedgerCOAKey(ledgerID), ledgerID, chart); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"accountCount": len(chart.Accounts)})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventChartUpdated, raw)
	return &chart, nil
}

// PostJournal validates and posts a double-entry voucher.
func (s *Service) PostJournal(ctx context.Context, ledgerID, userID string, j accounting.JournalEntry) (*accounting.JournalEntry, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	chart, err := s.GetChart(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err := accounting.PeriodFromDate(j.Date)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePeriodOpen(ctx, ledgerID, period); err != nil {
		return nil, err
	}
	j.Period = period
	j.LedgerID = ledgerID
	if err := accounting.ValidateJournal(&j, *chart); err != nil {
		return nil, err
	}
	id, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	j.ID = id
	j.PostedBy = userID
	j.PostedAt = time.Now().UTC()
	raw, _ := json.Marshal(j)
	ev, err := s.appendEvent(ctx, meta, userID, accounting.EventJournalPosted, raw)
	if err != nil {
		return nil, err
	}
	j.EventSeq = ev.Seq
	return &j, nil
}

// ListJournals loads journal entries from chain events.
func (s *Service) ListJournals(ctx context.Context, ledgerID, userID string) ([]accounting.JournalEntry, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	events, err := s.ListEvents(ctx, ledgerID, 1, meta.LatestSeq)
	if err != nil {
		return nil, err
	}
	var out []accounting.JournalEntry
	for _, ev := range events {
		if ev.Type != accounting.EventJournalPosted {
			continue
		}
		var j accounting.JournalEntry
		if json.Unmarshal(ev.Payload, &j) == nil {
			j.EventSeq = ev.Seq
			out = append(out, j)
		}
	}
	return out, nil
}

// ListPeriods returns period states (open months inferred from journals if missing).
func (s *Service) ListPeriods(ctx context.Context, ledgerID, userID string) ([]accounting.PeriodState, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerPeriodPrefix(ledgerID)+"%")
	if err != nil {
		return nil, err
	}
	byPeriod := map[string]accounting.PeriodState{}
	for _, row := range rows {
		var ps accounting.PeriodState
		if unmarshalStateValue(row.Value, &ps) == nil && ps.Period != "" {
			byPeriod[ps.Period] = ps
		}
	}
	journals, _ := s.ListJournals(ctx, ledgerID, userID)
	for _, j := range journals {
		if j.Period == "" {
			continue
		}
		if _, ok := byPeriod[j.Period]; !ok {
			byPeriod[j.Period] = accounting.PeriodState{Period: j.Period, Status: accounting.PeriodOpen}
		}
	}
	out := make([]accounting.PeriodState, 0, len(byPeriod))
	for _, ps := range byPeriod {
		out = append(out, ps)
	}
	// sort by period desc
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Period > out[i].Period {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

// ClosePeriod locks a month from new journals.
func (s *Service) ClosePeriod(ctx context.Context, ledgerID, userID, period string) (*accounting.PeriodState, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err = accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	ps := accounting.PeriodState{
		Period:   period,
		Status:   accounting.PeriodClosed,
		ClosedAt: now,
		ClosedBy: userID,
	}
	if err := s.putJSON(ctx, domain.LedgerPeriodKey(ledgerID, period), ledgerID, ps); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(ps)
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventPeriodClosed, raw)
	return &ps, nil
}

// ReopenPeriod opens a closed period (creator-level; any member for now).
func (s *Service) ReopenPeriod(ctx context.Context, ledgerID, userID, period string) (*accounting.PeriodState, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	period, err = accounting.NormalizePeriod(period)
	if err != nil {
		return nil, err
	}
	ps := accounting.PeriodState{Period: period, Status: accounting.PeriodOpen}
	if err := s.putJSON(ctx, domain.LedgerPeriodKey(ledgerID, period), ledgerID, ps); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(ps)
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventPeriodReopened, raw)
	return &ps, nil
}

// GetReports builds financial statements from posted journals.
func (s *Service) GetReports(ctx context.Context, ledgerID, userID, period string) (*accounting.FinancialReports, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	if period != "" {
		var err error
		period, err = accounting.NormalizePeriod(period)
		if err != nil {
			return nil, err
		}
	}
	chart, err := s.GetChart(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	journals, err := s.ListJournals(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	rep := accounting.BuildFinancialReports(*chart, journals, period)
	return &rep, nil
}

// LinkAttachment stores file on IPFS/disk and links to entry seq.
func (s *Service) LinkAttachment(
	ctx context.Context,
	ledgerID, userID string,
	entrySeq uint64,
	filename, mime string,
	size int64,
	body []byte,
	backup *storage.DualBackup,
) (*accounting.Attachment, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if entrySeq == 0 || entrySeq > meta.LatestSeq {
		return nil, domain.ErrEntryValidation
	}
	id, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	cid := ""
	ref := ""
	if backup != nil {
		putRes, err := backup.Put(ctx, ledgerID+":attach:"+id, "attach", body)
		if err == nil && putRes != nil {
			ref = putRes.Ref
			cid = putRes.IPFSCID
		}
	}
	att := accounting.Attachment{
		ID:         id,
		EntrySeq:   entrySeq,
		Filename:   filename,
		MimeType:   mime,
		Size:       size,
		CID:        cid,
		Ref:        ref,
		UploadedBy: userID,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.putJSON(ctx, domain.LedgerAttachmentKey(ledgerID, entrySeq, id), ledgerID, att); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(att)
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventAttachmentLinked, raw)
	return &att, nil
}

// ListAttachments returns all attachment metadata for a ledger.
func (s *Service) ListAttachments(ctx context.Context, ledgerID, userID string, entrySeq uint64) ([]accounting.Attachment, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	prefix := domain.LedgerAttachmentPrefix(ledgerID)
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		prefix+"%")
	if err != nil {
		return nil, err
	}
	var out []accounting.Attachment
	for _, row := range rows {
		var att accounting.Attachment
		if unmarshalStateValue(row.Value, &att) != nil {
			continue
		}
		if entrySeq > 0 && att.EntrySeq != entrySeq {
			continue
		}
		out = append(out, att)
	}
	return out, nil
}

// ImportBankStatement parses CSV and stores on chain.
func (s *Service) ImportBankStatement(ctx context.Context, ledgerID, userID, accountCode string, r io.Reader, filename string) (*accounting.BankStatement, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	stmt, err := accounting.ParseBankCSV(r, filename)
	if err != nil {
		return nil, err
	}
	stmt.LedgerID = ledgerID
	stmt.AccountCode = accountCode
	stmt.ImportedBy = userID
	stmt.ImportedAt = time.Now().UTC()
	if err := s.putJSON(ctx, domain.LedgerBankStmtKey(ledgerID, stmt.ID), ledgerID, stmt); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(stmt)
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventBankStatementImported, raw)
	return stmt, nil
}

// ListBankStatements returns imported statements.
func (s *Service) ListBankStatements(ctx context.Context, ledgerID, userID string) ([]accounting.BankStatement, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerBankStmtPrefix(ledgerID)+"%")
	if err != nil {
		return nil, err
	}
	var out []accounting.BankStatement
	for _, row := range rows {
		var stmt accounting.BankStatement
		if unmarshalStateValue(row.Value, &stmt) == nil {
			out = append(out, stmt)
		}
	}
	return out, nil
}

// MatchBankLine links a statement line to an entry seq.
func (s *Service) MatchBankLine(ctx context.Context, ledgerID, userID, stmtID, lineID string, entrySeq uint64) (*accounting.BankStatement, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	var stmt accounting.BankStatement
	if err := s.loadJSON(ctx, domain.LedgerBankStmtKey(ledgerID, stmtID), &stmt); err != nil {
		return nil, accounting.ErrStmtNotFound
	}
	found := false
	for i := range stmt.Lines {
		if stmt.Lines[i].ID == lineID {
			stmt.Lines[i].MatchedSeq = entrySeq
			found = true
			break
		}
	}
	if !found {
		return nil, accounting.ErrLineNotFound
	}
	if err := s.putJSON(ctx, domain.LedgerBankStmtKey(ledgerID, stmtID), ledgerID, stmt); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{
		"stmtId": stmtID, "lineId": lineID, "entrySeq": entrySeq,
	})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventReconciliationMatched, raw)
	return &stmt, nil
}

func (s *Service) ensurePeriodOpen(ctx context.Context, ledgerID, period string) error {
	var ps accounting.PeriodState
	if err := s.loadJSON(ctx, domain.LedgerPeriodKey(ledgerID, period), &ps); err != nil {
		return nil
	}
	if ps.Status == accounting.PeriodClosed {
		return accounting.ErrPeriodClosed
	}
	return nil
}

func (s *Service) loadJSON(ctx context.Context, key string, dest interface{}) error {
	rows, err := s.chain.Query(ctx, `SELECT key, value FROM world_state WHERE key = ?`, key)
	if err != nil || len(rows) == 0 {
		return domain.ErrLedgerNotFound
	}
	return unmarshalStateValue(rows[0].Value, dest)
}

func (s *Service) putJSON(ctx context.Context, key, ledgerID string, v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.submitOne(ctx, "state:"+key, ledgerID, miniledgerclient.TxRequest{Key: key, Value: raw})
}

// ImportBankStatementFromBytes helper for handlers.
func ImportBankStatementFromBytes(ctx context.Context, svc *Service, ledgerID, userID, accountCode string, data []byte, filename string) (*accounting.BankStatement, error) {
	return svc.ImportBankStatement(ctx, ledgerID, userID, accountCode, bytes.NewReader(data), filename)
}
