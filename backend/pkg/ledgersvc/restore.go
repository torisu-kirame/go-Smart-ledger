package ledgersvc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/chainstore"
)

var ErrRestoreConflict = errors.New("target ledger is not empty; use overwrite mode")

// RestoreOptions controls snapshot restore behavior.
type RestoreOptions struct {
	Overwrite bool
	SignerID  string
}

// RestoreSnapshot writes snapshot meta and events back to MiniLedger world state.
func (s *Service) RestoreSnapshot(ctx context.Context, targetLedgerID string, snap *LedgerSnapshot, opt RestoreOptions) error {
	if snap == nil || targetLedgerID == "" {
		return domain.ErrInvalidMember
	}
	var meta domain.LedgerMeta
	if err := json.Unmarshal(snap.Meta, &meta); err != nil {
		return err
	}
	if !opt.Overwrite {
		existing, err := s.loadMeta(ctx, targetLedgerID)
		if err == nil && existing.LatestSeq > 0 {
			return ErrRestoreConflict
		}
	}
	meta.ID = targetLedgerID
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, &meta); err != nil {
		return err
	}
	signer := opt.SignerID
	if signer == "" {
		signer = meta.CreatorID
	}
	for _, raw := range snap.Events {
		var ev domain.EventRecord
		if err := json.Unmarshal(raw, &ev); err != nil {
			continue
		}
		key := domain.LedgerEventKey(targetLedgerID, ev.Seq)
		if err := s.submitOne(ctx, "restore:"+key, targetLedgerID, chainstore.TxRequest{
			Key:   key,
			Value: raw,
		}); err != nil {
			return err
		}
	}
	// reload meta to ensure consistency
	loaded, err := s.loadMeta(ctx, targetLedgerID)
	if err != nil {
		return err
	}
	loaded.LatestSeq = meta.LatestSeq
	loaded.LatestRoot = meta.LatestRoot
	loaded.AnchorStatus = meta.AnchorStatus
	loaded.EntrySchema = meta.EntrySchema
	loaded.UpdatedAt = time.Now().UTC()
	return s.putMeta(ctx, loaded)
}
