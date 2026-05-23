package ledgersvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
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

// BatchImport appends valid rows, optional seal anchor, returns result.
func (s *Service) BatchImport(ctx context.Context, ledgerID, signerID string, rows []importxlsx.RowPreview, autoAnchor bool) (*ImportCommitResult, error) {
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
	schema := domain.ResolveEntrySchema(meta.EntrySchema)
	imported := 0
	skipped := 0
	for _, row := range rows {
		if row.Error != "" {
			skipped++
			continue
		}
		entry, err := importxlsx.ToEntry(row, schema)
		if err != nil {
			skipped++
			continue
		}
		sid, err := domain.SignerFromEntry(schema, entry.NormalizeData(), signerID)
		if err != nil {
			skipped++
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
		// reload meta after each append
		meta, err = s.loadMeta(ctx, ledgerID)
		if err != nil {
			return nil, err
		}
	}
	batchMeta, _ := json.Marshal(map[string]any{
		"imported": imported,
		"skipped":  skipped,
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
