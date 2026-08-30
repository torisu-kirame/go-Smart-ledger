package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const DefaultTableID = "default"

var (
	ErrMultiTableDisabled = errors.New("multi-table mode is disabled")
	ErrTableNotFound      = errors.New("ledger table not found")
	ErrInvalidTable       = errors.New("invalid ledger table")
	ErrTableHasEntries    = errors.New("table has entries and cannot be deleted")
)

// LedgerTable is one logical sheet inside a ledger (F49).
type LedgerTable struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	EntrySchema EntrySchema `json:"entrySchema"`
	SortOrder   int         `json:"sortOrder"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// NormalizeLedgerTables ensures Tables slice and EntrySchema stay aligned for legacy meta.
// When MultiTableEnabled and Tables is empty, the workbook has no sheets yet (user must create).
func NormalizeLedgerTables(meta *LedgerMeta) {
	if meta == nil {
		return
	}
	if IsProfessionalBookkeeping(meta) {
		meta.MultiTableEnabled = false
		meta.Tables = nil
		return
	}
	if meta.MultiTableEnabled && len(meta.Tables) == 0 {
		// Intentionally blank workbook: custom schema with no fields yet.
		if meta.EntrySchema.TemplateID == TemplateCustom && len(meta.EntrySchema.Fields) == 0 {
			return
		}
		if strings.TrimSpace(meta.EntrySchema.TemplateID) == "" && len(meta.EntrySchema.Fields) == 0 {
			meta.EntrySchema.TemplateID = TemplateCustom
			return
		}
		// Fall through: seed a first sheet from existing schema (legacy / templated).
	}
	schema := ResolveEntrySchema(meta.EntrySchema)
	if len(meta.Tables) == 0 {
		created := meta.CreatedAt
		if created.IsZero() {
			created = time.Now().UTC()
		}
		meta.Tables = []LedgerTable{{
			ID:          DefaultTableID,
			Name:        "默认",
			EntrySchema: schema,
			SortOrder:   0,
			CreatedAt:   created,
		}}
	}
	if t := TableByID(meta, DefaultTableID); t != nil {
		meta.EntrySchema = t.EntrySchema
	} else if len(meta.Tables) > 0 {
		meta.EntrySchema = meta.Tables[0].EntrySchema
	}
	if !meta.MultiTableEnabled {
		meta.EntrySchema = schema
		if len(meta.Tables) > 0 {
			meta.Tables[0].EntrySchema = schema
		}
	}
}

// TableByName finds a table by display name.
func TableByName(meta *LedgerMeta, name string) *LedgerTable {
	if meta == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	for i := range meta.Tables {
		if meta.Tables[i].Name == name {
			return &meta.Tables[i]
		}
	}
	return nil
}

// TableByID finds a table definition.
func TableByID(meta *LedgerMeta, tableID string) *LedgerTable {
	if meta == nil {
		return nil
	}
	tableID = strings.TrimSpace(tableID)
	if tableID == "" {
		tableID = DefaultTableID
	}
	for i := range meta.Tables {
		if meta.Tables[i].ID == tableID {
			return &meta.Tables[i]
		}
	}
	return nil
}

// ResolveTableID returns default when empty.
func ResolveTableID(meta *LedgerMeta, tableID string) string {
	tableID = strings.TrimSpace(tableID)
	if tableID == "" {
		return DefaultTableID
	}
	return tableID
}

// SchemaForTable returns entry schema for a table (or ledger-level when single-table mode).
func SchemaForTable(meta *LedgerMeta, tableID string) (EntrySchema, error) {
	NormalizeLedgerTables(meta)
	tableID = ResolveTableID(meta, tableID)
	if !meta.MultiTableEnabled {
		return ResolveEntrySchema(meta.EntrySchema), nil
	}
	t := TableByID(meta, tableID)
	if t == nil {
		return EntrySchema{}, ErrTableNotFound
	}
	return ResolveEntrySchema(t.EntrySchema), nil
}

// ValidateTableAccess checks table exists when multi-table is on.
func ValidateTableAccess(meta *LedgerMeta, tableID string) error {
	NormalizeLedgerTables(meta)
	tableID = ResolveTableID(meta, tableID)
	if !meta.MultiTableEnabled {
		return nil
	}
	if TableByID(meta, tableID) == nil {
		return ErrTableNotFound
	}
	return nil
}

// EntryTableIDFromPayload reads tableId from stored entry JSON.
func EntryTableIDFromPayload(raw []byte) string {
	var p EntryPayload
	if len(raw) == 0 {
		return DefaultTableID
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return DefaultTableID
	}
	if strings.TrimSpace(p.TableID) == "" {
		return DefaultTableID
	}
	return strings.TrimSpace(p.TableID)
}
