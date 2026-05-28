package ledgersvc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
)

// SetMultiTableEnabled toggles multi-table mode (creator only, simple ledgers).
func (s *Service) SetMultiTableEnabled(ctx context.Context, ledgerID, userID string, enabled bool) (*domain.LedgerMeta, error) {
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
	if meta.MultiTableEnabled == enabled {
		return meta, nil
	}
	if !enabled && len(meta.Tables) > 1 {
		return nil, domain.ErrInvalidTable
	}
	meta.MultiTableEnabled = enabled
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"enabled": enabled})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventMultiTableToggled, payload)
	return meta, nil
}

// CreateTable adds a table (requires multi-table mode).
func (s *Service) CreateTable(ctx context.Context, ledgerID, userID, name string, schema domain.EntrySchema) (*domain.LedgerMeta, error) {
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
		return nil, domain.ErrMultiTableDisabled
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidTable
	}
	for _, t := range meta.Tables {
		if t.Name == name {
			return nil, domain.ErrInvalidTable
		}
	}
	schema = domain.ResolveEntrySchema(schema)
	if err := domain.ValidateSchema(schema); err != nil {
		return nil, err
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
	meta.UpdatedAt = now
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"id": id, "name": name})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventTableCreated, payload)
	return meta, nil
}

// UpdateTable updates name and/or schema.
func (s *Service) UpdateTable(ctx context.Context, ledgerID, userID, tableID, name string, schema *domain.EntrySchema) (*domain.LedgerMeta, error) {
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
	t := domain.TableByID(meta, tableID)
	if t == nil {
		return nil, domain.ErrTableNotFound
	}
	idx := -1
	for i := range meta.Tables {
		if meta.Tables[i].ID == t.ID {
			idx = i
			break
		}
	}
	if name = strings.TrimSpace(name); name != "" {
		for _, other := range meta.Tables {
			if other.ID != t.ID && other.Name == name {
				return nil, domain.ErrInvalidTable
			}
		}
		meta.Tables[idx].Name = name
	}
	if schema != nil {
		sch := domain.ResolveEntrySchema(*schema)
		if err := domain.ValidateSchema(sch); err != nil {
			return nil, err
		}
		meta.Tables[idx].EntrySchema = sch
		if t.ID == domain.DefaultTableID {
			meta.EntrySchema = sch
		}
	}
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"id": t.ID, "name": meta.Tables[idx].Name})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventTableUpdated, payload)
	return meta, nil
}

// DeleteTable removes an empty non-default table.
func (s *Service) DeleteTable(ctx context.Context, ledgerID, userID, tableID string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	domain.NormalizeLedgerTables(meta)
	tableID = domain.ResolveTableID(meta, tableID)
	if tableID == domain.DefaultTableID {
		return nil, domain.ErrInvalidTable
	}
	if !meta.MultiTableEnabled {
		return nil, domain.ErrMultiTableDisabled
	}
	found := -1
	for i, t := range meta.Tables {
		if t.ID == tableID {
			found = i
			break
		}
	}
	if found < 0 {
		return nil, domain.ErrTableNotFound
	}
	if s.tableHasEntries(ctx, meta.ID, tableID) {
		return nil, domain.ErrTableHasEntries
	}
	meta.Tables = append(meta.Tables[:found], meta.Tables[found+1:]...)
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"id": tableID})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventTableDeleted, payload)
	return meta, nil
}

func (s *Service) tableHasEntries(ctx context.Context, ledgerID, tableID string) bool {
	events, err := s.ListEvents(ctx, ledgerID, 1, 0)
	if err != nil {
		return false
	}
	for _, ev := range events {
		if ev.Type != domain.EventEntryAdded {
			continue
		}
		if domain.EntryTableIDFromPayload(ev.Payload) == tableID {
			return true
		}
	}
	return false
}
