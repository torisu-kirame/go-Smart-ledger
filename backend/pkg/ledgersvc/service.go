package ledgersvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
)

// Service implements ledger business rules on top of Chainscore MiniLedger.
type Service struct {
	chain *miniledgerclient.Client
}

func New(chain *miniledgerclient.Client) *Service {
	return &Service{chain: chain}
}

func (s *Service) Online(ctx context.Context) bool {
	return s.chain.Ping(ctx) == nil
}

func (s *Service) Create(ctx context.Context, t domain.LedgerType, name, creatorID string, members []domain.Member) (*domain.LedgerMeta, error) {
	if err := domain.ValidateCreate(t, members); err != nil {
		return nil, err
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	meta := &domain.LedgerMeta{
		ID:           id,
		Type:         t,
		Name:         name,
		CreatorID:    creatorID,
		Members:      members,
		LatestSeq:    0,
		LatestRoot:   domain.MerkleRoot(nil),
		AnchorStatus: "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"name": name, "type": t, "members": members})
	if _, err := s.appendEvent(ctx, meta, creatorID, domain.EventLedgerCreated, payload); err != nil {
		return nil, err
	}
	return meta, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.LedgerMeta, error) {
	return s.loadMeta(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*domain.LedgerMeta, error) {
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerIndexPrefix())
	if err != nil {
		return nil, err
	}
	var out []*domain.LedgerMeta
	for _, row := range rows {
		var meta domain.LedgerMeta
		if err := json.Unmarshal(row.Value, &meta); err != nil {
			continue
		}
		if meta.ID != "" {
			out = append(out, &meta)
		}
	}
	return out, nil
}

func (s *Service) AppendEntry(ctx context.Context, ledgerID, signerID string, entry domain.EntryPayload) (*domain.EventRecord, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return nil, err
	}
	if err := domain.CanAppend(meta, signerID); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(entry)
	return s.appendEvent(ctx, meta, signerID, domain.EventEntryAdded, raw)
}

func (s *Service) ListEvents(ctx context.Context, ledgerID string, from, to uint64) ([]domain.EventRecord, error) {
	prefix := domain.LedgerEventKey(ledgerID, 0)
	_ = prefix
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerMetaKey(ledgerID)+":event:%")
	if err != nil {
		return nil, err
	}
	var out []domain.EventRecord
	for _, row := range rows {
		var ev domain.EventRecord
		if err := json.Unmarshal(row.Value, &ev); err != nil {
			continue
		}
		if ev.Seq >= from && (to == 0 || ev.Seq <= to) {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (s *Service) Anchor(ctx context.Context, ledgerID string, seqFrom, seqTo uint64) (txHash string, err error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return "", err
	}
	if seqTo == 0 {
		seqTo = meta.LatestSeq
	}
	if seqFrom == 0 {
		seqFrom = 1
	}
	seal, _ := json.Marshal(map[string]any{
		"ledgerId": ledgerID,
		"seqFrom":  seqFrom,
		"seqTo":    seqTo,
		"root":     meta.LatestRoot,
	})
	if _, err := s.appendEvent(ctx, meta, meta.CreatorID, domain.EventBatchSealed, seal); err != nil {
		return "", err
	}
	meta.AnchorStatus = "synced"
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return "", err
	}
	return "miniledger-anchored", nil
}

func (s *Service) Verify(ctx context.Context, ledgerID string) (bool, error) {
	meta, err := s.loadMeta(ctx, ledgerID)
	if err != nil {
		return false, err
	}
	events, err := s.ListEvents(ctx, ledgerID, 1, meta.LatestSeq)
	if err != nil {
		return false, err
	}
	hashes := make([]string, len(events))
	for i, e := range events {
		hashes[i] = e.Hash
	}
	return domain.MerkleRoot(hashes) == meta.LatestRoot, nil
}

func (s *Service) appendEvent(ctx context.Context, meta *domain.LedgerMeta, signerID, eventType string, payload []byte) (*domain.EventRecord, error) {
	seq := meta.LatestSeq + 1
	prev := meta.LatestRoot
	if seq > 1 {
		events, _ := s.ListEvents(ctx, meta.ID, seq-1, seq-1)
		if len(events) == 1 {
			prev = events[0].Hash
		}
	}
	hash := domain.EventHash(meta.ID, seq, eventType, payload, prev)
	ev := domain.EventRecord{
		Seq:       seq,
		Type:      eventType,
		Payload:   payload,
		PrevHash:  prev,
		Hash:      hash,
		SignerID:  signerID,
		CreatedAt: time.Now().UTC(),
	}
	raw, _ := json.Marshal(ev)
	if err := s.chain.Submit(ctx, miniledgerclient.TxRequest{
		Key:   domain.LedgerEventKey(meta.ID, seq),
		Value: raw,
	}); err != nil {
		return nil, err
	}
	var hashes []string
	if seq > 1 {
		prior, _ := s.ListEvents(ctx, meta.ID, 1, seq-1)
		hashes = make([]string, len(prior))
		for i, e := range prior {
			hashes[i] = e.Hash
		}
	}
	hashes = append(hashes, hash)
	meta.LatestSeq = seq
	meta.LatestRoot = domain.MerkleRoot(hashes)
	meta.UpdatedAt = time.Now().UTC()
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	return &ev, nil
}

func (s *Service) putMeta(ctx context.Context, meta *domain.LedgerMeta) error {
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return s.chain.Submit(ctx, miniledgerclient.TxRequest{
		Key:   domain.LedgerMetaKey(meta.ID),
		Value: raw,
	})
}

func (s *Service) loadMeta(ctx context.Context, id string) (*domain.LedgerMeta, error) {
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key = ?`,
		domain.LedgerMetaKey(id))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrLedgerNotFound
	}
	var meta domain.LedgerMeta
	if err := json.Unmarshal(rows[0].Value, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// MapDomainError converts domain errors to HTTP-friendly codes.
func MapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrLedgerNotFound):
		return 404
	case errors.Is(err, domain.ErrMultiNeedsTwo),
		errors.Is(err, domain.ErrPrivateOne),
		errors.Is(err, domain.ErrInvalidMember),
		errors.Is(err, domain.ErrUnauthorized):
		return 400
	default:
		return 500
	}
}
