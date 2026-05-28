package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/storage"
)

// LinkEntryAttachment stores a file linked to a chain event seq (F44, table-aware for F49).
func (s *Service) LinkEntryAttachment(
	ctx context.Context,
	ledgerID, userID, tableID string,
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
	tableID = domain.ResolveTableID(meta, tableID)
	if domain.IsProfessionalBookkeeping(meta) {
		tableID = ""
	} else {
		if err := domain.ValidateTableAccess(meta, tableID); err != nil {
			return nil, err
		}
		if err := s.validateAttachmentTarget(ctx, meta, tableID, entrySeq); err != nil {
			return nil, err
		}
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
	if domain.IsProfessionalBookkeeping(meta) {
		tableID = ""
	} else {
		tableID = domain.ResolveTableID(meta, tableID)
		if err := domain.ValidateTableAccess(meta, tableID); err != nil {
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
		if tableID != "" && domain.IsSimpleBookkeeping(meta) {
			attTable := att.TableID
			if attTable == "" {
				attTable = domain.DefaultTableID
			}
			if attTable != tableID {
				continue
			}
		}
		out = append(out, att)
	}
	return out, nil
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
