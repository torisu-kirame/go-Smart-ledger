package ledgersvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
)

var ErrImportHasErrors = errors.New("import contains invalid rows")

type ImportCommitResult struct {
	Imported   int    `json:"imported"`
	Skipped    int    `json:"skipped"`
	SeqTo      uint64 `json:"seqTo"`
	AnchorTx   string `json:"anchorTx,omitempty"`
	BackupRef  string `json:"backupRef,omitempty"`
	Root       string `json:"root,omitempty"`
}

// BatchImport appends valid rows into the given table, optional seal anchor.
func (s *Service) BatchImport(ctx context.Context, ledgerID, signerID, tableID string, rows []importxlsx.RowPreview, autoAnchor bool) (*ImportCommitResult, error) {
	valid := 0
	for _, r := range rows {
		if r.Error != "" {
			continue
		}
		valid++
	}
	if valid == 0 {
		return nil, ErrImportHasErrors
	}
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	tableID = strings.TrimSpace(tableID)
	var schema domain.EntrySchema
	for attempt := 0; attempt < 8; attempt++ {
		domain.NormalizeLedgerTables(meta)
		tid := domain.ResolveTableID(meta, tableID)
		if err := domain.ValidateTableAccess(meta, tid); err != nil {
			if !errors.Is(err, domain.ErrTableNotFound) || attempt == 7 {
				return nil, err
			}
			time.Sleep(time.Duration(attempt+1) * 15 * time.Millisecond)
			meta, err = s.loadMeta(ctx, ledgerID)
			if err != nil {
				return nil, err
			}
			continue
		}
		tableID = tid
		schema, err = domain.SchemaForTable(meta, tableID)
		if err != nil {
			return nil, err
		}
		break
	}
	presentKeys := importPresentKeys(rows)
	validateSchema := domain.RelaxRequiredForAbsentColumns(schema, presentKeys)

	imported := 0
	skipped := 0
	var firstSkip string
	for _, row := range rows {
		if row.Error != "" {
			skipped++
			if firstSkip == "" {
				firstSkip = row.Error
			}
			continue
		}
		row = fillImportUserFields(row, validateSchema, signerID)
		entry, err := importxlsx.ToEntry(row, validateSchema)
		if err != nil {
			skipped++
			if firstSkip == "" {
				firstSkip = err.Error()
			}
			continue
		}
		entry.TableID = tableID
		sid, err := domain.SignerFromEntry(schema, entry.NormalizeData(), signerID)
		if err != nil {
			skipped++
			if firstSkip == "" {
				firstSkip = err.Error()
			}
			continue
		}
		if err := domain.CanAppend(meta, sid); err != nil {
			return nil, err
		}
		raw, _ := json.Marshal(entry.ForChain(schema))
		if _, err := s.appendEvent(ctx, meta, sid, domain.EventEntryAdded, raw); err != nil {
			return nil, err
		}
		imported++
		// Do NOT loadMeta here: MiniLedger world_state can lag and rewind
		// LatestSeq, causing every row to overwrite the same event key.
		// appendEvent already updates meta.LatestSeq / LatestRoot in place.
	}
	if imported == 0 {
		msg := "all rows skipped"
		if firstSkip != "" {
			msg = fmt.Sprintf("all %d rows skipped: %s", skipped, firstSkip)
		}
		return nil, fmt.Errorf("%w: %s", ErrImportHasErrors, msg)
	}
	batchMeta, _ := json.Marshal(map[string]any{
		"imported": imported,
		"skipped":  skipped,
		"tableId":  tableID,
	})
	if _, err := s.appendEvent(ctx, meta, signerID, domain.EventImportBatch, batchMeta); err != nil {
		return nil, err
	}
	meta, _ = s.loadMeta(ctx, ledgerID)
	res := &ImportCommitResult{
		Imported: imported,
		Skipped:  skipped,
		SeqTo:    meta.LatestSeq,
		Root:     meta.LatestRoot,
	}
	if autoAnchor {
		tx, err := s.Anchor(ctx, ledgerID, 1, meta.LatestSeq)
		if err != nil {
			return res, fmt.Errorf("import ok but anchor failed: %w", err)
		}
		res.AnchorTx = tx
		meta, _ = s.loadMeta(ctx, ledgerID)
		res.Root = meta.LatestRoot
	}
	return res, nil
}

func importPresentKeys(rows []importxlsx.RowPreview) map[string]bool {
	out := map[string]bool{}
	for _, r := range rows {
		if r.Error != "" {
			continue
		}
		for k := range r.Cells {
			if strings.TrimSpace(k) != "" {
				out[k] = true
			}
		}
		for _, k := range []string{"date", "type", "amount", "category", "note", "counterparty"} {
			var v string
			switch k {
			case "date":
				v = r.Date
			case "type":
				v = r.Type
			case "amount":
				v = r.Amount
			case "category":
				v = r.Category
			case "note":
				v = r.Note
			case "counterparty":
				v = r.Counterparty
			}
			if strings.TrimSpace(v) != "" {
				out[k] = true
			}
		}
	}
	return out
}

func fillImportUserFields(row importxlsx.RowPreview, schema domain.EntrySchema, signerID string) importxlsx.RowPreview {
	signerID = strings.TrimSpace(signerID)
	if signerID == "" {
		return row
	}
	schema = domain.ResolveEntrySchema(schema)
	cells := row.Cells
	if cells == nil {
		cells = map[string]string{}
	} else {
		cp := make(map[string]string, len(cells)+2)
		for k, v := range cells {
			cp[k] = v
		}
		cells = cp
	}
	changed := false
	for _, f := range schema.Fields {
		if f.Type != domain.FieldUser {
			continue
		}
		if strings.TrimSpace(cells[f.Key]) == "" {
			cells[f.Key] = signerID
			changed = true
		}
	}
	if !changed {
		return row
	}
	row.Cells = cells
	return row
}
