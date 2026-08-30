package ledgersvc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/chainstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/storage"
)

var ErrAttachmentNotFound = errors.New("attachment not found")

// LinkAttachment is the public API used by HTTP handlers.
func (s *Service) LinkAttachment(
	ctx context.Context,
	ledgerID, userID, tableID string,
	entrySeq uint64,
	filename, mime string,
	size int64,
	body []byte,
	aux *accounting.AuxiliaryDims,
	backup *storage.DualBackup,
) (*accounting.Attachment, error) {
	return s.LinkEntryAttachment(ctx, ledgerID, userID, tableID, entrySeq, filename, mime, size, body, aux, backup)
}

// ListAttachments returns attachment metadata (optional tableId / entrySeq filters).
func (s *Service) ListAttachments(ctx context.Context, ledgerID, userID, tableID string, entrySeq uint64) ([]accounting.Attachment, error) {
	return s.ListEntryAttachments(ctx, ledgerID, userID, tableID, entrySeq)
}

// LinkEntryAttachment stores a file linked to a chain event seq (F44, table-aware for F49).
func (s *Service) LinkEntryAttachment(
	ctx context.Context,
	ledgerID, userID, tableID string,
	entrySeq uint64,
	filename, mime string,
	size int64,
	body []byte,
	aux *accounting.AuxiliaryDims,
	backup *storage.DualBackup,
) (*accounting.Attachment, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if entrySeq == 0 || entrySeq > meta.LatestSeq {
		return nil, domain.ErrEntryValidation
	}
	tableID = domain.ResolveTableID(meta, tableID)
	if err := domain.ValidateTableAccess(meta, tableID); err != nil {
		return nil, err
	}
	if err := s.validateAttachmentTarget(ctx, meta, tableID, entrySeq); err != nil {
		return nil, err
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
		TableID:    tableID,
		EntrySeq:   entrySeq,
		Filename:   filename,
		MimeType:   mime,
		Size:       size,
		CID:        cid,
		Ref:        ref,
		Auxiliary:  normalizeAux(aux),
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

// ListEntryAttachments lists attachments, optionally filtered by tableId and entrySeq.
func (s *Service) ListEntryAttachments(ctx context.Context, ledgerID, userID, tableID string, entrySeq uint64) ([]accounting.Attachment, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	filterTable := strings.TrimSpace(tableID)
	if filterTable != "" {
		filterTable = domain.ResolveTableID(meta, filterTable)
		if err := domain.ValidateTableAccess(meta, filterTable); err != nil {
			return nil, err
		}
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
		if filterTable != "" {
			attTable := att.TableID
			if attTable == "" {
				attTable = domain.DefaultTableID
			}
			if attTable != filterTable {
				continue
			}
		}
		out = append(out, att)
	}
	return out, nil
}

// UpdateAttachmentAuxiliary sets department/project/counterparty tags (F41).
func (s *Service) UpdateAttachmentAuxiliary(
	ctx context.Context,
	ledgerID, userID, attachID string,
	aux accounting.AuxiliaryDims,
) (*accounting.Attachment, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	att, err := s.loadAttachmentByID(ctx, ledgerID, attachID)
	if err != nil {
		return nil, err
	}
	att.Auxiliary = normalizeAux(&aux)
	if err := s.putJSON(ctx, domain.LedgerAttachmentKey(ledgerID, att.EntrySeq, att.ID), ledgerID, att); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{
		"attachmentId": att.ID,
		"entrySeq":     att.EntrySeq,
		"tableId":      att.TableID,
		"auxiliary":    att.Auxiliary,
	})
	_, _ = s.appendEvent(ctx, meta, userID, accounting.EventAttachmentAuxUpdated, payload)
	return &att, nil
}

func (s *Service) loadAttachmentByID(ctx context.Context, ledgerID, attachID string) (accounting.Attachment, error) {
	attachID = strings.TrimSpace(attachID)
	if attachID == "" {
		return accounting.Attachment{}, ErrAttachmentNotFound
	}
	prefix := domain.LedgerAttachmentPrefix(ledgerID)
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		prefix+"%")
	if err != nil {
		return accounting.Attachment{}, err
	}
	for _, row := range rows {
		var att accounting.Attachment
		if unmarshalStateValue(row.Value, &att) != nil {
			continue
		}
		if att.ID == attachID {
			return att, nil
		}
	}
	return accounting.Attachment{}, ErrAttachmentNotFound
}

func normalizeAux(aux *accounting.AuxiliaryDims) *accounting.AuxiliaryDims {
	if aux == nil {
		return nil
	}
	out := accounting.AuxiliaryDims{
		Department:   strings.TrimSpace(aux.Department),
		Project:      strings.TrimSpace(aux.Project),
		Counterparty: strings.TrimSpace(aux.Counterparty),
	}
	if out.Department == "" && out.Project == "" && out.Counterparty == "" {
		return nil
	}
	return &out
}

func (s *Service) validateAttachmentTarget(ctx context.Context, meta *domain.LedgerMeta, tableID string, entrySeq uint64) error {
	events, err := s.ListEvents(ctx, meta.ID, entrySeq, entrySeq)
	if err != nil || len(events) == 0 {
		return domain.ErrEntryValidation
	}
	ev := events[0]
	if ev.Type != domain.EventEntryAdded {
		return domain.ErrEntryValidation
	}
	got := domain.EntryTableIDFromPayload(ev.Payload)
	if got != domain.ResolveTableID(meta, tableID) {
		return domain.ErrEntryValidation
	}
	return nil
}

func (s *Service) putJSON(ctx context.Context, key, ledgerID string, v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.submitOne(ctx, "state:"+key, ledgerID, chainstore.TxRequest{Key: key, Value: raw})
}
