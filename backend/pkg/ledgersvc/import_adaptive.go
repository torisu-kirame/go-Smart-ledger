package ledgersvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
)

// AdaptiveImportCommitResult includes new table metadata.
type AdaptiveImportCommitResult struct {
	Import      *ImportCommitResult `json:"import"`
	TableID     string              `json:"tableId"`
	TableName   string              `json:"tableName"`
	EntrySchema domain.EntrySchema  `json:"entrySchema"`
}

// ImportAdaptiveCommit enables multi-table, creates a table from inferred schema, and imports rows.
func (s *Service) ImportAdaptiveCommit(
	ctx context.Context,
	ledgerID, userID, signerID string,
	tableName string,
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
	if domain.IsProfessionalBookkeeping(meta) {
		return nil, domain.ErrBookkeepingModeMismatch
	}
	domain.NormalizeLedgerTables(meta)
	if !meta.MultiTableEnabled {
		meta, err = s.SetMultiTableEnabled(ctx, ledgerID, userID, true)
		if err != nil {
			return nil, err
		}
	}
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		tableName = "导入数据"
	}
	tableName = uniqueTableName(meta, tableName)
	schema = domain.ResolveEntrySchema(schema)
	if err := domain.ValidateSchema(schema); err != nil {
		return nil, err
	}
	meta, err = s.createTableInternal(ctx, meta, userID, tableName, schema)
	if err != nil {
		return nil, err
	}
	t := domain.TableByName(meta, tableName)
	if t == nil {
		return nil, domain.ErrTableNotFound
	}
	res, err := s.BatchImport(ctx, ledgerID, signerID, t.ID, rows, autoAnchor)
	if err != nil {
		return nil, err
	}
	return &AdaptiveImportCommitResult{
		Import:      res,
		TableID:     t.ID,
		TableName:   t.Name,
		EntrySchema: t.EntrySchema,
	}, nil
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

func (s *Service) createTableInternal(ctx context.Context, meta *domain.LedgerMeta, userID, name string, schema domain.EntrySchema) (*domain.LedgerMeta, error) {
	for _, t := range meta.Tables {
		if t.Name == name {
			return nil, domain.ErrInvalidTable
		}
	}
	id, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	meta.Tables = append(meta.Tables, domain.LedgerTable{
		ID:          id,
		Name:        name,
		EntrySchema: schema,
		SortOrder:   len(meta.Tables),
		CreatedAt:   now,
	})
	meta.MultiTableEnabled = true
	meta.UpdatedAt = now
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	return s.loadMeta(ctx, meta.ID)
}
