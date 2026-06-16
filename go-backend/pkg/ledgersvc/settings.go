package ledgersvc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
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

// SetApprovalPolicy updates multi-ledger approval settings (creator only).
func (s *Service) SetApprovalPolicy(ctx context.Context, ledgerID, userID string, ap domain.ApprovalPolicy) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	if meta.Type != domain.LedgerMulti {
		return nil, domain.ErrInvalidApprovalPolicy
	}
	memberCount := len(meta.Members)
	if ap.Enabled {
		if memberCount < 2 {
			return nil, domain.ErrInvalidApprovalPolicy
		}
		ap.Threshold = memberCount
	} else {
		ap.Threshold = 1
	}
	meta.ApprovalPolicy = ap
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(ap)
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventApprovalPolicyUpdated, payload)
	return meta, nil
}

// EnableEncryption turns on group E2E for a ledger that has no encryption yet (creator only).
func (s *Service) EnableEncryption(ctx context.Context, ledgerID, userID string, enc domain.LedgerEncryption) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	if meta.Encryption.Enabled {
		return nil, domain.ErrEncryptionAlreadyEnabled
	}
	if !enc.Enabled || len(enc.WrappedKeys) == 0 {
		return nil, domain.ErrInvalidMember
	}
	for _, m := range meta.Members {
		if enc.WrappedKeys[m.ID] == "" {
			return nil, domain.ErrInvalidMember
		}
	}
	if enc.Algo == "" {
		enc.Algo = "aes-gcm-v1"
	}
	meta.Encryption = enc
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{
		"enabled": true, "algo": enc.Algo, "memberIds": memberIDs(meta.Members),
	})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventEncryptionEnabled, payload)
	return meta, nil
}

// SetPassphraseViewPolicy toggles whether members may register login-wrapped passphrase copies (creator only).
func (s *Service) SetPassphraseViewPolicy(ctx context.Context, ledgerID, userID string, enabled bool) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	if !meta.Encryption.Enabled {
		return nil, domain.ErrInvalidMember
	}
	meta.Encryption.PassphraseViewEnabled = enabled
	if !enabled && meta.Encryption.PassphraseWrappedKeys != nil {
		meta.Encryption.PassphraseWrappedKeys = nil
	}
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"passphraseViewEnabled": enabled})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventEncryptionEnabled, payload)
	return meta, nil
}

// RegisterPassphraseViewWrap stores the current user's login-password-wrapped ledger passphrase.
func (s *Service) RegisterPassphraseViewWrap(ctx context.Context, ledgerID, userID, wrapped string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if !meta.Encryption.Enabled {
		return nil, domain.ErrInvalidMember
	}
	if !meta.Encryption.PassphraseViewEnabled && meta.CreatorID != userID {
		return nil, domain.ErrUnauthorized
	}
	if wrapped == "" {
		return nil, domain.ErrInvalidMember
	}
	if meta.Encryption.PassphraseWrappedKeys == nil {
		meta.Encryption.PassphraseWrappedKeys = map[string]string{}
	}
	meta.Encryption.PassphraseWrappedKeys[userID] = wrapped
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func memberIDs(members []domain.Member) []string {
	out := make([]string, len(members))
	for i, m := range members {
		out[i] = m.ID
	}
	return out
}
