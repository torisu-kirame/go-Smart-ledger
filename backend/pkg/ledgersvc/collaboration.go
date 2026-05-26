package ledgersvc

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
)

// CreateOptions optional policies for new ledgers (F17/F19).
type CreateOptions struct {
	BookkeepingMode string
	ApprovalPolicy  domain.ApprovalPolicy
	Encryption      domain.LedgerEncryption
	StorageLocation string
}

func (s *Service) ListForUser(ctx context.Context, userID string) ([]*domain.LedgerMeta, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	if userID == "" {
		return all, nil
	}
	out := make([]*domain.LedgerMeta, 0, len(all))
	for _, m := range all {
		if !m.ArchivedAt.IsZero() {
			continue
		}
		if domain.IsMember(m, userID) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *Service) GetForUser(ctx context.Context, ledgerID, userID string) (*domain.LedgerMeta, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	if !meta.ArchivedAt.IsZero() {
		return nil, domain.ErrLedgerNotFound
	}
	if userID != "" && !domain.IsMember(meta, userID) {
		return nil, domain.ErrUnauthorized
	}
	return meta, nil
}

// ProposeEntry creates a pending entry when approval is required; otherwise appends directly.
func (s *Service) ProposeEntry(ctx context.Context, ledgerID, actorID string, entry domain.EntryPayload) (*domain.PendingEntry, *domain.EventRecord, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, nil, err
	}
	if err := domain.CanAppend(meta, actorID); err != nil {
		return nil, nil, err
	}
	if domain.IsProfessionalBookkeeping(meta) {
		return nil, nil, domain.ErrBookkeepingModeMismatch
	}
	schema := domain.ResolveEntrySchema(meta.EntrySchema)
	data := entry.NormalizeData()
	if err := domain.ValidateEntryData(schema, data); err != nil {
		return nil, nil, err
	}
	signerID, err := domain.SignerFromEntry(schema, data, actorID)
	if err != nil {
		return nil, nil, err
	}
	raw, _ := json.Marshal(entry.ForChain(schema))

	if !domain.ApprovalRequired(meta) {
		ev, err := s.appendEvent(ctx, meta, signerID, domain.EventEntryAdded, raw)
		return nil, ev, err
	}

	pid, err := snowflake.NextString()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	pending := &domain.PendingEntry{
		ID:         pid,
		LedgerID:   ledgerID,
		ProposerID: signerID,
		Payload:    raw,
		Approvals:  []string{},
		Status:     domain.PendingStatusPending,
		CreatedAt:  now,
	}
	if err := s.putPending(ctx, pending); err != nil {
		return nil, nil, err
	}
	prop, _ := json.Marshal(map[string]any{
		"pendingId": pid, "proposerId": signerID,
	})
	if _, err := s.appendEvent(ctx, meta, signerID, domain.EventEntryProposed, prop); err != nil {
		return nil, nil, err
	}
	return pending, nil, nil
}

func (s *Service) ListPending(ctx context.Context, ledgerID, userID string) ([]domain.PendingEntry, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	_ = meta
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerPendingPrefix(ledgerID)+"%")
	if err != nil {
		return nil, err
	}
	var out []domain.PendingEntry
	for _, row := range rows {
		var p domain.PendingEntry
		if err := unmarshalStateValue(row.Value, &p); err != nil {
			continue
		}
		if p.Status == domain.PendingStatusPending {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Service) ApprovePending(ctx context.Context, ledgerID, pendingID, approverID string) (*domain.EventRecord, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	if !domain.IsMember(meta, approverID) {
		return nil, domain.ErrUnauthorized
	}
	p, err := s.loadPending(ctx, ledgerID, pendingID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.PendingStatusPending {
		return nil, domain.ErrInvalidApproval
	}
	if p.ProposerID == approverID {
		return nil, domain.ErrCannotApproveOwn
	}
	if domain.HasApproved(p, approverID) {
		return nil, domain.ErrInvalidApproval
	}
	p.Approvals = append(p.Approvals, approverID)
	th := meta.ApprovalPolicy.Threshold
	if th < 1 {
		th = 2
	}
	if len(p.Approvals) < th {
		if err := s.putPending(ctx, p); err != nil {
			return nil, err
		}
		ap, _ := json.Marshal(map[string]any{"pendingId": pendingID, "approverId": approverID})
		_, _ = s.appendEvent(ctx, meta, approverID, domain.EventEntryApproved, ap)
		return nil, nil
	}
	ev, err := s.appendEvent(ctx, meta, p.ProposerID, domain.EventEntryAdded, p.Payload)
	if err != nil {
		return nil, err
	}
	p.Status = domain.PendingStatusApproved
	_ = s.deleteKey(ctx, domain.LedgerPendingKey(ledgerID, pendingID))
	return ev, nil
}

func (s *Service) RejectPending(ctx context.Context, ledgerID, pendingID, approverID string) error {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return err
	}
	if !domain.IsMember(meta, approverID) {
		return domain.ErrUnauthorized
	}
	p, err := s.loadPending(ctx, ledgerID, pendingID)
	if err != nil {
		return err
	}
	if p.Status != domain.PendingStatusPending {
		return domain.ErrInvalidApproval
	}
	p.Status = domain.PendingStatusRejected
	_ = s.deleteKey(ctx, domain.LedgerPendingKey(ledgerID, pendingID))
	rej, _ := json.Marshal(map[string]any{"pendingId": pendingID, "approverId": approverID})
	_, err = s.appendEvent(ctx, meta, approverID, domain.EventEntryRejected, rej)
	return err
}

func (s *Service) InviteMember(ctx context.Context, ledgerID, inviterID, inviteeID, role string) (*domain.MemberInvite, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	if meta.Type != domain.LedgerMulti {
		return nil, domain.ErrInvalidMember
	}
	if !domain.IsMember(meta, inviterID) {
		return nil, domain.ErrUnauthorized
	}
	if inviteeID == "" {
		return nil, domain.ErrInvalidMember
	}
	if inviterID == inviteeID {
		return nil, domain.ErrCannotInviteSelf
	}
	if domain.IsMember(meta, inviteeID) {
		return nil, domain.ErrAlreadyMember
	}
	if existing, err := s.loadInvite(ctx, ledgerID, inviteeID); err == nil && existing.Status == domain.InviteStatusPending {
		return nil, domain.ErrInviteAlreadyPending
	}
	inv := &domain.MemberInvite{
		LedgerID:  ledgerID,
		InviteeID: inviteeID,
		InviterID: inviterID,
		Role:      role,
		Status:    domain.InviteStatusPending,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.putInvite(ctx, inv); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{
		"inviteeId": inviteeID, "inviterId": inviterID, "role": role,
	})
	_, _ = s.appendEvent(ctx, meta, inviterID, domain.EventMemberInvited, payload)
	return inv, nil
}

func (s *Service) ListInvitesForUser(ctx context.Context, userID string) ([]domain.MemberInvite, error) {
	suffix := domain.LedgerInviteSuffix(userID)
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		"%"+suffix)
	if err != nil {
		return nil, err
	}
	var out []domain.MemberInvite
	for _, row := range rows {
		var inv domain.MemberInvite
		if err := unmarshalStateValue(row.Value, &inv); err != nil {
			continue
		}
		if inv.Status == domain.InviteStatusPending {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (s *Service) ListInvites(ctx context.Context, ledgerID, userID string) ([]domain.MemberInvite, error) {
	if _, err := s.GetForUser(ctx, ledgerID, userID); err != nil {
		return nil, err
	}
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerMetaKey(ledgerID)+":invite:%")
	if err != nil {
		return nil, err
	}
	var out []domain.MemberInvite
	for _, row := range rows {
		var inv domain.MemberInvite
		if err := unmarshalStateValue(row.Value, &inv); err != nil {
			continue
		}
		if inv.Status == domain.InviteStatusPending {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (s *Service) AcceptInvite(ctx context.Context, ledgerID, userID string) (*domain.LedgerMeta, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	inv, err := s.loadInvite(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	if inv.InviteeID != userID {
		return nil, domain.ErrUnauthorized
	}
	if domain.IsMember(meta, userID) {
		return nil, domain.ErrAlreadyMember
	}
	role := inv.Role
	if role == "" {
		role = "member"
	}
	idInt, _ := strconv.ParseInt(ledgerID, 10, 64)
	members := append(meta.Members, domain.Member{ID: userID, Role: role})
	members = assignMemberAddresses(s.hd, idInt, members)
	meta.Members = members
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	_ = s.deleteKey(ctx, domain.LedgerInviteKey(ledgerID, userID))
	payload, _ := json.Marshal(map[string]any{"userId": userID, "role": role})
	_, _ = s.appendEvent(ctx, meta, userID, domain.EventMemberJoined, payload)
	return meta, nil
}

// SetStorageLocation updates where ledger snapshots are preferred to be stored.
func (s *Service) SetStorageLocation(ctx context.Context, ledgerID, userID, location string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	loc := domain.NormalizeStorageLocation(location)
	meta.StorageLocation = loc
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// SyncEvents returns events with seq > sinceSeq (F18 incremental sync).
func (s *Service) SyncEvents(ctx context.Context, ledgerID, userID string, sinceSeq uint64) ([]domain.EventRecord, *domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, nil, err
	}
	events, err := s.ListEvents(ctx, ledgerID, sinceSeq+1, meta.LatestSeq)
	if err != nil {
		return nil, nil, err
	}
	return events, meta, nil
}

// RotateGroupKeys updates wrapped keys after a new member joins (F19).
func (s *Service) RotateGroupKeys(ctx context.Context, ledgerID, actorID string, wrapped map[string]string) (*domain.LedgerMeta, error) {
	meta, err := s.GetForUser(ctx, ledgerID, actorID)
	if err != nil {
		return nil, err
	}
	if !meta.Encryption.Enabled {
		return nil, domain.ErrInvalidMember
	}
	if meta.Encryption.WrappedKeys == nil {
		meta.Encryption.WrappedKeys = map[string]string{}
	}
	for k, v := range wrapped {
		meta.Encryption.WrappedKeys[k] = v
	}
	meta.Encryption.Algo = "aes-gcm-v1"
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"wrappedKeys": wrapped})
	_, _ = s.appendEvent(ctx, meta, actorID, domain.EventGroupKeyRotated, payload)
	return meta, nil
}

func (s *Service) putPending(ctx context.Context, p *domain.PendingEntry) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.submitOne(ctx, "pending:"+p.ID, p.LedgerID, miniledgerclient.TxRequest{
		Key:   domain.LedgerPendingKey(p.LedgerID, p.ID),
		Value: raw,
	})
}

func (s *Service) loadPending(ctx context.Context, ledgerID, pendingID string) (*domain.PendingEntry, error) {
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key = ?`,
		domain.LedgerPendingKey(ledgerID, pendingID))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrPendingNotFound
	}
	var p domain.PendingEntry
	if err := unmarshalStateValue(rows[0].Value, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) putInvite(ctx context.Context, inv *domain.MemberInvite) error {
	raw, err := json.Marshal(inv)
	if err != nil {
		return err
	}
	return s.submitOne(ctx, "invite:"+inv.InviteeID, inv.LedgerID, miniledgerclient.TxRequest{
		Key:   domain.LedgerInviteKey(inv.LedgerID, inv.InviteeID),
		Value: raw,
	})
}

func (s *Service) loadInvite(ctx context.Context, ledgerID, inviteeID string) (*domain.MemberInvite, error) {
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key = ?`,
		domain.LedgerInviteKey(ledgerID, inviteeID))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrInviteNotFound
	}
	var inv domain.MemberInvite
	if err := unmarshalStateValue(rows[0].Value, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *Service) deleteKey(ctx context.Context, key string) error {
	return s.submitOne(ctx, "delete:"+key, "", miniledgerclient.TxRequest{Key: key, Value: json.RawMessage(`null`)})
}
