package ledgersvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importfile"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
)

// SheetCSVImportResult is the response for one-shot CSV → sheet import.
type SheetCSVImportResult struct {
	Mode        string                 `json:"mode"` // created | appended
	TableID     string                 `json:"tableId"`
	TableName   string                 `json:"tableName"`
	EntrySchema domain.EntrySchema     `json:"entrySchema"`
	Import      *ImportCommitResult    `json:"import"`
	ParsedRows  int                    `json:"parsedRows"`
	Valid       int                    `json:"valid"`
	Invalid     int                    `json:"invalid"`
}

// ImportSheetCSV imports CSV into a ledger sheet.
// - tableID empty: enable multi-table if needed, create a new sheet from CSV headers, import all rows.
// - tableID set: append rows to that sheet (matched by column header ↔ field label/key).
func (s *Service) ImportSheetCSV(
	ctx context.Context,
	ledgerID, userID, signerID string,
	csvData []byte,
	tableID, sheetName string,
	autoAnchor bool,
) (*SheetCSVImportResult, error) {
	if len(csvData) == 0 {
		return nil, fmt.Errorf("%w: empty csv", ErrImportHasErrors)
	}
	signerID = strings.TrimSpace(signerID)
	if signerID == "" {
		signerID = userID
	}
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if domain.IsProfessionalBookkeeping(meta) {
		return nil, domain.ErrBookkeepingModeMismatch
	}
	domain.NormalizeLedgerTables(meta)

	tableID = strings.TrimSpace(tableID)
	if tableID == "" {
		return s.importSheetCSVCreate(ctx, meta, ledgerID, userID, signerID, csvData, sheetName, autoAnchor)
	}
	return s.importSheetCSVAppend(ctx, meta, ledgerID, userID, signerID, csvData, tableID, autoAnchor)
}

func (s *Service) importSheetCSVCreate(
	ctx context.Context,
	meta *domain.LedgerMeta,
	ledgerID, userID, signerID string,
	csvData []byte,
	sheetName string,
	autoAnchor bool,
) (*SheetCSVImportResult, error) {
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	schema, rows, err := importfile.PrepareCSVForNewSheet(csvData)
	if err != nil {
		return nil, err
	}
	valid, invalid := countRowValidity(rows)
	if valid == 0 {
		return nil, ErrImportHasErrors
	}
	if !meta.MultiTableEnabled {
		meta, err = s.SetMultiTableEnabled(ctx, ledgerID, userID, true)
		if err != nil {
			return nil, err
		}
	}
	sheetName = strings.TrimSpace(sheetName)
	if sheetName == "" {
		sheetName = "导入明细"
	}
	// Same name exists → merge missing fields and append (do not create duplicate sheet)
	if existing := domain.TableByName(meta, sheetName); existing != nil {
		return s.importSheetCSVAppend(ctx, meta, ledgerID, userID, signerID, csvData, existing.ID, autoAnchor)
	}
	meta, err = s.createTableInternal(ctx, meta, userID, sheetName, schema)
	if err != nil {
		return nil, err
	}
	t := domain.TableByName(meta, sheetName)
	if t == nil {
		return nil, domain.ErrTableNotFound
	}
	res, err := s.BatchImport(ctx, ledgerID, signerID, t.ID, rows, autoAnchor)
	if err != nil {
		return nil, err
	}
	return &SheetCSVImportResult{
		Mode:        "created",
		TableID:     t.ID,
		TableName:   t.Name,
		EntrySchema: t.EntrySchema,
		Import:      res,
		ParsedRows:  len(rows),
		Valid:       valid,
		Invalid:     invalid,
	}, nil
}

func (s *Service) importSheetCSVAppend(
	ctx context.Context,
	meta *domain.LedgerMeta,
	ledgerID, userID, signerID string,
	csvData []byte,
	tableID string,
	autoAnchor bool,
) (*SheetCSVImportResult, error) {
	if err := domain.CanAppend(meta, userID); err != nil {
		return nil, err
	}
	tableID = domain.ResolveTableID(meta, tableID)
	if err := domain.ValidateTableAccess(meta, tableID); err != nil {
		return nil, err
	}
	t := domain.TableByID(meta, tableID)
	if t == nil {
		return nil, domain.ErrTableNotFound
	}
	schema, err := domain.SchemaForTable(meta, tableID)
	if err != nil {
		return nil, err
	}
	headers, rawRows, err := importfile.ParseCSVBytes(csvData)
	if err != nil {
		return nil, err
	}
	headers, rawRows = importfile.FilterSeqColumns(headers, rawRows)
	incoming, err := importfile.BuildImportSchemaFromHeaders(headers)
	if err != nil {
		return nil, err
	}
	merged, _ := domain.MergeEntrySchema(schema, incoming)
	if len(merged.Fields) > len(domain.ResolveEntrySchema(schema).Fields) {
		if meta.CreatorID != userID {
			return nil, domain.ErrUnauthorized
		}
		meta, err = s.UpdateTable(ctx, ledgerID, userID, tableID, "", &merged)
		if err != nil {
			return nil, err
		}
		t = domain.TableByID(meta, tableID)
		if t == nil {
			return nil, domain.ErrTableNotFound
		}
		schema = merged
	}
	rows := importfile.MapRowsByHeaders(headers, rawRows, schema)
	if len(rows) == 0 {
		return nil, importxlsx.ErrNoDataRows
	}
	valid, invalid := countRowValidity(rows)
	if valid == 0 {
		return nil, ErrImportHasErrors
	}
	res, err := s.BatchImport(ctx, ledgerID, signerID, tableID, rows, autoAnchor)
	if err != nil {
		return nil, err
	}
	return &SheetCSVImportResult{
		Mode:        "appended",
		TableID:     tableID,
		TableName:   t.Name,
		EntrySchema: schema,
		Import:      res,
		ParsedRows:  len(rows),
		Valid:       valid,
		Invalid:     invalid,
	}, nil
}

func countRowValidity(rows []importxlsx.RowPreview) (valid, invalid int) {
	for _, r := range rows {
		if r.Error == "" {
			valid++
		} else {
			invalid++
		}
	}
	return
}
