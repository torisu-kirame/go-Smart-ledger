package ledgersvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

// SheetEditCommit is a batch of sheet mutations applied atomically as chain events.
type SheetEditCommit struct {
	TableID     string              `json:"tableId"`
	EntrySchema *domain.EntrySchema `json:"entrySchema,omitempty"`
	VoidSeqs    []uint64            `json:"voidSeqs,omitempty"`
	NewRows     []map[string]any    `json:"newRows,omitempty"`
	RowOrder    []uint64            `json:"rowOrder,omitempty"` // 0 = next new row seq in order
	SignerID    string              `json:"signerId,omitempty"`
}

// ReorderTables persists sheet tab order and emits TablesReordered.
func (s *Service) ReorderTables(ctx context.Context, ledgerID, userID string, orderedIDs []string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	domain.NormalizeLedgerTables(meta)
	if !meta.MultiTableEnabled {
		return nil, domain.ErrMultiTableDisabled
	}
	if len(orderedIDs) == 0 {
		return nil, domain.ErrInvalidTable
	}
	byID := map[string]domain.LedgerTable{}
	for _, t := range meta.Tables {
		byID[t.ID] = t
	}
	if len(orderedIDs) != len(meta.Tables) {
		return nil, domain.ErrInvalidTable
	}
	next := make([]domain.LedgerTable, 0, len(orderedIDs))
	seen := map[string]bool{}
	for i, id := range orderedIDs {
		id = strings.TrimSpace(id)
		t, ok := byID[id]
		if !ok || seen[id] {
			return nil, domain.ErrTableNotFound
		}
		seen[id] = true
		t.SortOrder = i
		next = append(next, t)
	}
	meta.Tables = next
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"tableIds": orderedIDs})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventTablesReordered, payload)
	return meta, nil
}

// ReorderEntries sets display order for a sheet and emits EntriesReordered.
func (s *Service) ReorderEntries(ctx context.Context, ledgerID, userID, tableID string, order []uint64) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if err := domain.CanAppend(meta, userID); err != nil {
		return nil, err
	}
	domain.NormalizeLedgerTables(meta)
	tableID = domain.ResolveTableID(meta, tableID)
	if err := domain.ValidateTableAccess(meta, tableID); err != nil {
		return nil, err
	}
	return s.setTableRowOrder(ctx, meta, userID, tableID, order)
}

// CommitSheetEdit applies schema/void/append/reorder and records each op on chain.
func (s *Service) CommitSheetEdit(ctx context.Context, ledgerID, userID string, edit SheetEditCommit) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if err := domain.CanAppend(meta, userID); err != nil {
		return nil, err
	}
	domain.NormalizeLedgerTables(meta)
	tableID := domain.ResolveTableID(meta, edit.TableID)
	if err := domain.ValidateTableAccess(meta, tableID); err != nil {
		return nil, err
	}
	signerID := strings.TrimSpace(edit.SignerID)
	if signerID == "" {
		signerID = userID
	}

	ops := 0
	if edit.EntrySchema != nil {
		if meta.CreatorID != userID {
			return nil, domain.ErrUnauthorized
		}
		sch := domain.ResolveEntrySchema(*edit.EntrySchema)
		if err := domain.ValidateSchema(sch); err != nil {
			return nil, err
		}
		meta, err = s.UpdateTable(ctx, ledgerID, userID, tableID, "", &sch)
		if err != nil {
			return nil, err
		}
		ops++
	}

	voided := map[uint64]bool{}
	for _, seq := range edit.VoidSeqs {
		if seq == 0 {
			continue
		}
		if err := s.voidEntryUnlocked(ctx, meta, userID, tableID, seq); err != nil {
			return nil, err
		}
		voided[seq] = true
		ops++
	}

	schema, err := domain.SchemaForTable(meta, tableID)
	if err != nil {
		return nil, err
	}
	newSeqs := make([]uint64, 0, len(edit.NewRows))
	for _, data := range edit.NewRows {
		entry := domain.EntryPayload{TableID: tableID, Data: anyMapToStrings(data)}
		rawData := entry.NormalizeData()
		if err := domain.ValidateEntryData(schema, rawData); err != nil {
			return nil, err
		}
		sid, err := domain.SignerFromEntry(schema, rawData, signerID)
		if err != nil {
			return nil, err
		}
		raw, _ := json.Marshal(entry.ForChain(schema))
		ev, err := s.appendEvent(ctx, meta, sid, domain.EventEntryAdded, raw)
		if err != nil {
			return nil, err
		}
		newSeqs = append(newSeqs, ev.Seq)
		ops++
	}

	rowOrder := make([]uint64, 0, len(edit.RowOrder))
	ni := 0
	for _, seq := range edit.RowOrder {
		if seq == 0 {
			if ni < len(newSeqs) {
				rowOrder = append(rowOrder, newSeqs[ni])
				ni++
			}
			continue
		}
		if voided[seq] {
			continue
		}
		rowOrder = append(rowOrder, seq)
	}
	for ni < len(newSeqs) {
		rowOrder = append(rowOrder, newSeqs[ni])
		ni++
	}

	if len(rowOrder) > 0 || len(edit.RowOrder) > 0 || len(voided) > 0 {
		meta, err = s.setTableRowOrder(ctx, meta, userID, tableID, rowOrder)
		if err != nil {
			return nil, err
		}
		ops++
	}

	summary, _ := json.Marshal(map[string]any{
		"tableId":  tableID,
		"voided":   len(voided),
		"appended": len(newSeqs),
		"rowOrder": rowOrder,
		"ops":      ops,
		"schema":   edit.EntrySchema != nil,
	})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventSheetEditCommitted, summary)
	return meta, nil
}

func (s *Service) voidEntryUnlocked(ctx context.Context, meta *domain.LedgerMeta, userID, tableID string, seq uint64) error {
	events, err := s.ListEvents(ctx, meta.ID, seq, seq)
	if err != nil {
		return err
	}
	if len(events) != 1 || events[0].Type != domain.EventEntryAdded {
		return domain.ErrEntryValidation
	}
	if domain.EntryTableIDFromPayload(events[0].Payload) != tableID {
		return domain.ErrEntryValidation
	}
	payload, _ := json.Marshal(map[string]any{"seq": seq, "tableId": tableID})
	_, err = s.appendEvent(ctx, meta, userID, domain.EventEntryVoided, payload)
	return err
}

func (s *Service) setTableRowOrder(ctx context.Context, meta *domain.LedgerMeta, userID, tableID string, order []uint64) (*domain.LedgerMeta, error) {
	domain.NormalizeLedgerTables(meta)
	idx := -1
	for i := range meta.Tables {
		if meta.Tables[i].ID == tableID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, domain.ErrTableNotFound
	}
	meta.Tables[idx].RowOrder = append([]uint64(nil), order...)
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"tableId": tableID, "rowOrder": order})
	_, err := s.appendEvent(ctx, meta, userID, domain.EventEntriesReordered, payload)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// CollectVoidedSeqs returns seqs voided via EntryVoided events.
func CollectVoidedSeqs(events []domain.EventRecord) map[uint64]bool {
	out := map[uint64]bool{}
	for _, ev := range events {
		if ev.Type != domain.EventEntryVoided {
			continue
		}
		var p struct {
			Seq uint64 `json:"seq"`
		}
		raw := ev.Payload
		if len(raw) > 0 && raw[0] == '"' {
			var s string
			if json.Unmarshal(raw, &s) == nil {
				_ = json.Unmarshal([]byte(s), &p)
			}
		} else {
			_ = json.Unmarshal(raw, &p)
		}
		if p.Seq == 0 {
			var m map[string]any
			if json.Unmarshal(raw, &m) == nil {
				switch v := m["seq"].(type) {
				case float64:
					p.Seq = uint64(v)
				case json.Number:
					n, _ := v.Int64()
					p.Seq = uint64(n)
				}
			}
		}
		if p.Seq > 0 {
			out[p.Seq] = true
		}
	}
	return out
}

func anyMapToStrings(m map[string]any) map[string]string {
	out := map[string]string{}
	if m == nil {
		return out
	}
	for k, v := range m {
		if v == nil {
			out[k] = ""
			continue
		}
		switch t := v.(type) {
		case string:
			out[k] = t
		case float64:
			if t == float64(int64(t)) {
				out[k] = fmt.Sprintf("%d", int64(t))
			} else {
				out[k] = fmt.Sprint(t)
			}
		case bool:
			if t {
				out[k] = "true"
			} else {
				out[k] = "false"
			}
		default:
			out[k] = fmt.Sprint(v)
		}
	}
	return out
}
