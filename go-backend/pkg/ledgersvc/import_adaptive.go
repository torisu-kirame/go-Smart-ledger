package ledgersvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
)

// AdaptiveImportCommitResult includes new table metadata.
type AdaptiveImportCommitResult struct {
	Import      *ImportCommitResult `json:"import"`
	TableID     string              `json:"tableId"`
	TableName   string              `json:"tableName"`
	EntrySchema domain.EntrySchema  `json:"entrySchema"`
	Mode        string              `json:"mode"` // created | appended
}

// ImportAdaptiveCommit enables multi-table, then either appends to an existing
// sheet (tableID) or creates a new sheet (tableName). Missing fields are merged
// when appending.
func (s *Service) ImportAdaptiveCommit(
	ctx context.Context,
	ledgerID, userID, signerID string,
	tableID, tableName string,
	schema domain.EntrySchema,
	rows []importxlsx.RowPreview,
	autoAnchor bool,
) (*AdaptiveImportCommitResult, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	domain.NormalizeLedgerTables(meta)
	if !meta.MultiTableEnabled {
		meta, err = s.SetMultiTableEnabled(ctx, ledgerID, userID, true)
		if err != nil {
			return nil, err
		}
	}
	schema = domain.ResolveEntrySchema(schema)
	if err := domain.ValidateSchema(schema); err != nil {
		return nil, err
	}

	tableID = strings.TrimSpace(tableID)
	if tableID != "" {
		existing := domain.TableByID(meta, tableID)
		if existing == nil {
			return nil, domain.ErrTableNotFound
		}
		return s.importAdaptiveAppend(ctx, meta, ledgerID, userID, signerID, existing, schema, rows, autoAnchor)
	}

	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		tableName = "导入数据"
	}
	createName := uniqueTableName(meta, tableName)
	meta, newTable, err := s.createTableInternal(ctx, meta, userID, createName, schema)
	if err != nil {
		return nil, err
	}
	res, err := s.BatchImport(ctx, ledgerID, signerID, newTable.ID, rows, autoAnchor)
	if err != nil {
		return nil, err
	}
	return &AdaptiveImportCommitResult{
		Import:      res,
		TableID:     newTable.ID,
		TableName:   newTable.Name,
		EntrySchema: newTable.EntrySchema,
		Mode:        "created",
	}, nil
}

func (s *Service) importAdaptiveAppend(
	ctx context.Context,
	meta *domain.LedgerMeta,
	ledgerID, userID, signerID string,
	existing *domain.LedgerTable,
	incoming domain.EntrySchema,
	rows []importxlsx.RowPreview,
	autoAnchor bool,
) (*AdaptiveImportCommitResult, error) {
	merged, remap := domain.MergeEntrySchema(existing.EntrySchema, incoming)
	if len(merged.Fields) != len(domain.ResolveEntrySchema(existing.EntrySchema).Fields) {
		var err error
		meta, err = s.UpdateTable(ctx, ledgerID, userID, existing.ID, "", &merged)
		if err != nil {
			return nil, err
		}
		existing = domain.TableByID(meta, existing.ID)
		if existing == nil {
			return nil, domain.ErrTableNotFound
		}
	}
	mapped := remapImportRows(rows, remap)
	res, err := s.BatchImport(ctx, ledgerID, signerID, existing.ID, mapped, autoAnchor)
	if err != nil {
		return nil, err
	}
	return &AdaptiveImportCommitResult{
		Import:      res,
		TableID:     existing.ID,
		TableName:   existing.Name,
		EntrySchema: existing.EntrySchema,
		Mode:        "appended",
	}, nil
}

func remapImportRows(rows []importxlsx.RowPreview, remap map[string]string) []importxlsx.RowPreview {
	if len(remap) == 0 {
		return rows
	}
	out := make([]importxlsx.RowPreview, len(rows))
	for i, r := range rows {
		nr := r
		if len(r.Cells) == 0 {
			out[i] = nr
			continue
		}
		cells := make(map[string]string, len(r.Cells))
		for k, v := range r.Cells {
			if dest, ok := remap[k]; ok {
				cells[dest] = v
			} else {
				cells[k] = v
			}
		}
		nr.Cells = cells
		out[i] = nr
	}
	return out
}

func uniqueTableName(meta *domain.LedgerMeta, base string) string {
	used := map[string]bool{}
	for _, t := range meta.Tables {
		used[t.Name] = true
	}
	if !used[base] {
		return base
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s (%d)", base, i)
		if !used[candidate] {
			return candidate
		}
	}
	return base + " (新)"
}

func (s *Service) createTableInternal(ctx context.Context, meta *domain.LedgerMeta, userID, name string, schema domain.EntrySchema) (*domain.LedgerMeta, *domain.LedgerTable, error) {
	for _, t := range meta.Tables {
		if t.Name == name {
			return nil, nil, domain.ErrInvalidTable
		}
	}
	id, err := snowflake.NextString()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	newTable := domain.LedgerTable{
		ID:          id,
		Name:        name,
		EntrySchema: schema,
		SortOrder:   len(meta.Tables),
		CreatedAt:   now,
	}
	meta.Tables = append(meta.Tables, newTable)
	meta.MultiTableEnabled = true
	meta.UpdatedAt = now
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, nil, err
	}
	// Prefer in-memory meta: MiniLedger world_state reads can lag right after putMeta.
	return meta, &meta.Tables[len(meta.Tables)-1], nil
}
