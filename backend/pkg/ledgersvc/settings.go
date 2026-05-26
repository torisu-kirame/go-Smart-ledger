package ledgersvc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

// UpdateLedger applies creator-only name changes for an active ledger.
func (s *Service) UpdateLedger(ctx context.Context, ledgerID, userID, name string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidMember
	}
	if meta.Name == name {
		return meta, nil
	}
	meta.Name = name
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"name": name})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventLedgerUpdated, payload)
	return meta, nil
}

// ArchiveLedger soft-deletes a ledger for all members (creator only).
func (s *Service) ArchiveLedger(ctx context.Context, ledgerID, userID string) error {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return err
	}
	if meta.CreatorID != userID {
		return domain.ErrUnauthorized
	}
	if !meta.ArchivedAt.IsZero() {
		return nil
	}
	now := time.Now().UTC()
	meta.ArchivedAt = now
	meta.UpdatedAt = now
	if err := s.putMeta(ctx, meta); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"archivedAt": now.Format(time.RFC3339)})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventLedgerArchived, payload)
	return nil
}
