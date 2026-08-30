package ledgersvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/chainstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/evmanchor"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ledgerhd"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/txqueue"
)

// Service implements ledger business rules on top of Chainscore MiniLedger.
type Service struct {
	chain    chainstore.Store
	hd       *ledgerhd.Deriver
	queue    *txqueue.Queue
	external evmanchor.Anchorer
	// ledgerLocks serializes append/meta writes per ledger so concurrent
	// EntryAdded requests cannot claim the same seq (last-write-wins race).
	ledgerLocks sync.Map // ledgerID -> *sync.Mutex
}

// ErrSeqConflict means another writer already occupied LatestSeq+1.
var ErrSeqConflict = errors.New("ledger sequence conflict")

func New(chain chainstore.Store, hd *ledgerhd.Deriver, queue *txqueue.Queue, external evmanchor.Anchorer) *Service {
	if external == nil {
		external = evmanchor.Noop()
	}
	return &Service{chain: chain, hd: hd, queue: queue, external: external}
}

func (s *Service) ledgerLock(ledgerID string) *sync.Mutex {
	v, _ := s.ledgerLocks.LoadOrStore(ledgerID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (s *Service) Online(ctx context.Context) bool {
	return s.chain.Ping(ctx) == nil
}

func (s *Service) Create(ctx context.Context, t domain.LedgerType, name, creatorID string, members []domain.Member, schema domain.EntrySchema, opts CreateOptions) (*domain.LedgerMeta, error) {
	if err := domain.ValidateCreate(t, members); err != nil {
		return nil, err
	}
	mode := domain.NormalizeBookkeepingMode(opts.BookkeepingMode)
	blankSheets := false
	// Simple ledgers: sheet workbook — no preset template; sheets created by user.
	if strings.TrimSpace(schema.TemplateID) == "" || schema.TemplateID == domain.TemplateCustom {
		if len(schema.Fields) == 0 {
			schema = domain.EntrySchema{TemplateID: domain.TemplateCustom, Fields: nil}
			blankSheets = true
		} else {
			schema.TemplateID = domain.TemplateCustom
			if err := domain.ValidateSchema(schema); err != nil {
				return nil, err
			}
		}
	} else {
		schema = domain.ResolveEntrySchema(schema)
		if err := domain.ValidateSchema(schema); err != nil {
			return nil, err
		}
	}
	idInt, err := snowflake.NextInt64()
	if err != nil {
		return nil, err
	}
	id := strconv.FormatInt(idInt, 10)
	members = assignMemberAddresses(s.hd, idInt, members)
	ledgerAddr := ""
	if s.hd != nil {
		idx := ledgerIndexFromID(idInt)
		ledgerAddr, err = s.hd.LedgerAddress(idx)
		if err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	ap := opts.ApprovalPolicy
	if !ap.Enabled && ap.Threshold == 0 {
		ap = domain.DefaultApprovalPolicy(t, len(members))
	}
	enc := opts.Encryption
	if enc.Enabled && enc.Algo == "" {
		enc.Algo = "aes-gcm-v1"
	}
	meta := &domain.LedgerMeta{
		ID:                id,
		Type:              t,
		Name:              name,
		CreatorID:         creatorID,
		LedgerAddress:     ledgerAddr,
		Members:           members,
		BookkeepingMode:   mode,
		EntrySchema:       schema,
		ApprovalPolicy:    ap,
		Encryption:        enc,
		StorageLocation:   domain.NormalizeStorageLocation(opts.StorageLocation),
		LatestSeq:         0,
		LatestRoot:        domain.MerkleRoot(nil),
		AnchorStatus:      "pending",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	meta.MultiTableEnabled = true
	if blankSheets {
		meta.Tables = []domain.LedgerTable{}
	}
	domain.NormalizeLedgerTables(meta)
	if err := s.putMeta(ctx, meta); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{
		"name": name, "type": t, "members": members, "ledgerAddress": ledgerAddr,
		"bookkeepingMode": mode, "entrySchema": schema,
	})
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
		`SELECT key, value FROM world_state WHERE key LIKE ? AND key NOT LIKE ? ORDER BY key`,
		domain.LedgerIndexPrefix(), domain.LedgerIndexPrefix()+":event:%")
	if err != nil {
		return nil, err
	}
	var out []*domain.LedgerMeta
	for _, row := range rows {
		var meta domain.LedgerMeta
		if err := unmarshalStateValue(row.Value, &meta); err != nil {
			continue
		}
		if meta.ID != "" {
			out = append(out, &meta)
		}
	}
	return out, nil
}

func (s *Service) AppendEntry(ctx context.Context, ledgerID, signerID string, entry domain.EntryPayload) (*domain.EventRecord, error) {
	_, ev, err := s.ProposeEntry(ctx, ledgerID, signerID, entry)
	return ev, err
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
		if err := unmarshalStateValue(row.Value, &ev); err != nil {
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
	txHash = "miniledger-anchored"
	if s.external != nil && s.external.Enabled() {
		ext, err := s.external.Anchor(ctx, ledgerID, meta.LatestRoot, seqFrom, seqTo)
		if err != nil {
			return "", fmt.Errorf("miniledger sealed but external anchor failed: %w", err)
		}
		if ext != nil {
			meta.ExternalAnchor = &domain.ExternalAnchorRecord{
				TxHash:      ext.TxHash,
				ChainID:     ext.ChainID,
				ChainName:   ext.ChainName,
				ExplorerURL: ext.ExplorerURL,
				MerkleRoot:  meta.LatestRoot,
				SeqFrom:     seqFrom,
				SeqTo:       seqTo,
				AnchoredAt:  ext.AnchoredAt,
			}
			meta.UpdatedAt = time.Now().UTC()
			if err := s.putMeta(ctx, meta); err != nil {
				return "", err
			}
			extPayload, _ := json.Marshal(map[string]any{
				"txHash": ext.TxHash, "chainId": ext.ChainID, "root": meta.LatestRoot,
			})
			_, _ = s.appendEvent(ctx, meta, meta.CreatorID, domain.EventExternalAnchored, extPayload)
			txHash = ext.TxHash
		}
	}
	return txHash, nil
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
	if meta == nil || strings.TrimSpace(meta.ID) == "" {
		return nil, domain.ErrLedgerNotFound
	}
	var last error
	for attempt := 0; attempt < 8; attempt++ {
		ev, err := s.appendEventLocked(ctx, meta, signerID, eventType, payload)
		if err == nil {
			return ev, nil
		}
		if !errors.Is(err, ErrSeqConflict) {
			return nil, err
		}
		last = err
		time.Sleep(time.Duration(attempt+1) * 5 * time.Millisecond)
	}
	if last == nil {
		last = ErrSeqConflict
	}
	return nil, last
}

func (s *Service) appendEventLocked(ctx context.Context, meta *domain.LedgerMeta, signerID, eventType string, payload []byte) (*domain.EventRecord, error) {
	mu := s.ledgerLock(meta.ID)
	mu.Lock()
	defer mu.Unlock()

	fresh, err := s.loadMeta(ctx, meta.ID)
	if err != nil {
		return nil, err
	}
	// Never rewind seq from a lagging world_state read — BatchImport used to
	// loadMeta after each row and reuse the same seq, overwriting EntryAdded.
	if meta.LatestSeq > fresh.LatestSeq {
		fresh.LatestSeq = meta.LatestSeq
		fresh.LatestRoot = meta.LatestRoot
	}
	if maxSeq := s.maxEventSeq(ctx, meta.ID); maxSeq > fresh.LatestSeq {
		fresh.LatestSeq = maxSeq
	}
	// Preserve meta mutations applied on the caller pointer when the
	// world_state read is briefly stale after a prior putMeta.
	if !meta.ArchivedAt.IsZero() {
		fresh.ArchivedAt = meta.ArchivedAt
	}
	callerNewer := !meta.UpdatedAt.IsZero() && (fresh.UpdatedAt.IsZero() || !meta.UpdatedAt.Before(fresh.UpdatedAt))
	if name := strings.TrimSpace(meta.Name); name != "" && name != fresh.Name && callerNewer {
		fresh.Name = name
	}
	if callerNewer {
		// Multi-table / sheet edits must not be wiped by a stale reload
		// (import-to-new-table: putMeta then appendEvent).
		fresh.MultiTableEnabled = meta.MultiTableEnabled
		if meta.Tables != nil {
			fresh.Tables = append([]domain.LedgerTable(nil), meta.Tables...)
		}
		fresh.EntrySchema = meta.EntrySchema
	}
	ev, err := s.appendEventUsingMetaUnlocked(ctx, fresh, signerID, eventType, payload)
	if err != nil {
		return nil, err
	}
	*meta = *fresh
	return ev, nil
}

func (s *Service) maxEventSeq(ctx context.Context, ledgerID string) uint64 {
	events, err := s.ListEvents(ctx, ledgerID, 1, 0)
	if err != nil || len(events) == 0 {
		return 0
	}
	var max uint64
	for _, e := range events {
		if e.Seq > max {
			max = e.Seq
		}
	}
	return max
}

// appendEventUsingMetaUnlocked appends using meta.LatestSeq as-is.
// Caller MUST hold ledgerLock(meta.ID).
func (s *Service) appendEventUsingMetaUnlocked(ctx context.Context, meta *domain.LedgerMeta, signerID, eventType string, payload []byte) (*domain.EventRecord, error) {
	seq := meta.LatestSeq + 1
	if existing, _ := s.ListEvents(ctx, meta.ID, seq, seq); len(existing) > 0 {
		return nil, fmt.Errorf("%w: seq %d already exists", ErrSeqConflict, seq)
	}
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
	eventTx := chainstore.TxRequest{
		Key:   domain.LedgerEventKey(meta.ID, seq),
		Value: raw,
	}
	if err := s.submitOne(ctx, fmt.Sprintf("event:%s:%d", meta.ID, seq), meta.ID, eventTx); err != nil {
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
	if err := s.putMetaUnlocked(ctx, meta); err != nil {
		return nil, err
	}
	return &ev, nil
}

func (s *Service) putMeta(ctx context.Context, meta *domain.LedgerMeta) error {
	if meta == nil || strings.TrimSpace(meta.ID) == "" {
		return domain.ErrLedgerNotFound
	}
	mu := s.ledgerLock(meta.ID)
	mu.Lock()
	defer mu.Unlock()
	return s.putMetaUnlocked(ctx, meta)
}

func (s *Service) putMetaUnlocked(ctx context.Context, meta *domain.LedgerMeta) error {
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return s.submitOne(ctx, "meta:"+meta.ID, meta.ID, chainstore.TxRequest{
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
	if err := unmarshalStateValue(rows[0].Value, &meta); err != nil {
		return nil, err
	}
	domain.NormalizeLedgerTables(&meta)
	return &meta, nil
}

// unmarshalStateValue decodes MiniLedger world_state values (object or JSON string).
func unmarshalStateValue(raw json.RawMessage, dest interface{}) error {
	if err := json.Unmarshal(raw, dest); err == nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), dest)
}

func ledgerIndexFromID(id int64) uint32 {
	return uint32(id & 0x7fffffff)
}

func assignMemberAddresses(hd *ledgerhd.Deriver, ledgerID int64, members []domain.Member) []domain.Member {
	if hd == nil {
		return members
	}
	idx := ledgerIndexFromID(ledgerID)
	out := make([]domain.Member, len(members))
	for i, m := range members {
		out[i] = m
		if m.Address == "" {
			addr, err := hd.MemberAddress(idx, uint32(i))
			if err == nil {
				out[i].Address = ledgerhd.NormalizeAddress(addr)
			}
		} else {
			out[i].Address = ledgerhd.NormalizeAddress(m.Address)
		}
	}
	return out
}

// MapDomainError converts domain errors to HTTP-friendly codes.
func MapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrLedgerNotFound):
		return 404
	case errors.Is(err, domain.ErrMultiNeedsTwo),
		errors.Is(err, domain.ErrPrivateOne),
		errors.Is(err, domain.ErrInvalidMember),
		errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrEntryValidation),
		errors.Is(err, domain.ErrInvalidSchema),
		errors.Is(err, domain.ErrPendingNotFound),
		errors.Is(err, domain.ErrInviteNotFound),
		errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrInvalidApproval),
		errors.Is(err, domain.ErrCannotApproveOwn),
		errors.Is(err, ErrRestoreConflict),
		errors.Is(err, domain.ErrEncryptionAlreadyEnabled),
		errors.Is(err, domain.ErrInvalidApprovalPolicy),
		errors.Is(err, domain.ErrBookkeepingModeMismatch):
		return 400
	default:
		return 500
	}
}
