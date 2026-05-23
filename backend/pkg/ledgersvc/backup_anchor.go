package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

// BackupAnchorResult is returned after dual backup + chain CID record.
type BackupAnchorResult struct {
	Ref       string `json:"ref"`
	IPFSCID   string `json:"ipfsCid,omitempty"`
	Seq       uint64 `json:"seq"`
	Anchored  bool   `json:"anchoredOnChain"`
}

// RecordBackupAnchor appends BackupAnchored event and updates ledger meta.
func (s *Service) RecordBackupAnchor(ctx context.Context, ledgerID, signerID, ref, ipfsCID string) (*BackupAnchorResult, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{
		"ref":       ref,
		"ipfsCid":   ipfsCID,
		"root":      meta.LatestRoot,
		"seq":       meta.LatestSeq,
		"storedAt":  time.Now().UTC().Format(time.RFC3339),
	})
	ev, err := s.appendEvent(ctx, meta, signerID, domain.EventBackupAnchored, payload)
	if err != nil {
		return nil, err
	}
	meta, _ = s.loadMeta(ctx, ledgerID)
	meta.LastBackupRef = ref
	meta.LastBackupCID = ipfsCID
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	return &BackupAnchorResult{
		Ref:      ref,
		IPFSCID:  ipfsCID,
		Seq:      ev.Seq,
		Anchored: true,
	}, nil
}
