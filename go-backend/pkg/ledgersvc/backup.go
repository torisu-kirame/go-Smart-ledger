package ledgersvc

import (
	"context"
	"encoding/json"
	"time"
)

// LedgerSnapshot is encrypted backup payload.
type LedgerSnapshot struct {
	LedgerID  string                `json:"ledgerId"`
	ExportedAt time.Time            `json:"exportedAt"`
	Meta      json.RawMessage       `json:"meta"`
	Events    []json.RawMessage     `json:"events"`
}

// BuildSnapshot serializes ledger meta and all events for backup/restore preview.
func (s *Service) BuildSnapshot(ctx context.Context, ledgerID string) (*LedgerSnapshot, []byte, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, nil, err
	}
	events, err := s.ListEvents(ctx, ledgerID, 1, meta.LatestSeq)
	if err != nil {
		return nil, nil, err
	}
	metaRaw, _ := json.Marshal(meta)
	evRaws := make([]json.RawMessage, len(events))
	for i, e := range events {
		evRaws[i], _ = json.Marshal(e)
	}
	snap := &LedgerSnapshot{
		LedgerID:   ledgerID,
		ExportedAt: time.Now().UTC(),
		Meta:       metaRaw,
		Events:     evRaws,
	}
	raw, err := json.Marshal(snap)
	return snap, raw, err
}
