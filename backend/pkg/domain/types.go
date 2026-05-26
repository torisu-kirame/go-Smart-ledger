package domain

import "time"

type LedgerType string

const (
	LedgerPrivate LedgerType = "private"
	LedgerMulti   LedgerType = "multi"
)

// Ledger data storage location (user-selectable).
const (
	StorageLocationCloud = "cloud"
	StorageLocationLocal = "local"
	StorageLocationIPFS  = "ipfs"
)

func NormalizeStorageLocation(s string) string {
	switch s {
	case StorageLocationCloud, StorageLocationLocal, StorageLocationIPFS:
		return s
	default:
		return StorageLocationIPFS
	}
}

type Member struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Role    string `json:"role,omitempty"`
}

type LedgerMeta struct {
	ID             string           `json:"id"`
	Type           LedgerType       `json:"type"`
	Name           string           `json:"name"`
	CreatorID      string           `json:"creatorId"`
	LedgerAddress  string           `json:"ledgerAddress,omitempty"`
	Members        []Member         `json:"members"`
	EntrySchema    EntrySchema      `json:"entrySchema,omitempty"`
	ApprovalPolicy ApprovalPolicy   `json:"approvalPolicy,omitempty"`
	Encryption     LedgerEncryption `json:"encryption,omitempty"`
	LatestSeq      uint64           `json:"latestSeq"`
	LatestRoot     string           `json:"latestRoot"`
	AnchorStatus    string                `json:"anchorStatus"`
	ExternalAnchor  *ExternalAnchorRecord `json:"externalAnchor,omitempty"`
	LastBackupRef     string                `json:"lastBackupRef,omitempty"`
	LastBackupCID     string                `json:"lastBackupCid,omitempty"`
	StorageLocation   string                `json:"storageLocation,omitempty"`
	ArchivedAt        time.Time             `json:"archivedAt,omitempty"`
	CreatedAt         time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

// ExternalAnchorRecord stores an EVM/L2 anchor transaction (F29).
type ExternalAnchorRecord struct {
	TxHash      string    `json:"txHash"`
	ChainID     uint64    `json:"chainId"`
	ChainName   string    `json:"chainName,omitempty"`
	ExplorerURL string    `json:"explorerUrl,omitempty"`
	MerkleRoot  string    `json:"merkleRoot,omitempty"`
	SeqFrom     uint64    `json:"seqFrom,omitempty"`
	SeqTo       uint64    `json:"seqTo,omitempty"`
	AnchoredAt  time.Time `json:"anchoredAt"`
}

type EventRecord struct {
	Seq       uint64    `json:"seq"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	PrevHash  string    `json:"prevHash"`
	Hash      string    `json:"hash"`
	SignerID  string    `json:"signerId"`
	CreatedAt time.Time `json:"createdAt"`
}

// EntryPayload is stored on chain for EntryAdded events.
// Prefer Data + schemaId; legacy top-level fields are still read for old events.
type EntryPayload struct {
	SchemaID string            `json:"schemaId,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Date         string `json:"date,omitempty"`
	Type         string `json:"type,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Category     string `json:"category,omitempty"`
	Note         string `json:"note,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
}

// NormalizeData fills Data from legacy fields when missing.
func (e *EntryPayload) NormalizeData() map[string]string {
	if len(e.Data) > 0 {
		return e.Data
	}
	out := map[string]string{}
	if e.Date != "" {
		out["date"] = e.Date
	}
	if e.Type != "" {
		out["type"] = e.Type
	}
	if e.Amount != "" {
		out["amount"] = e.Amount
	}
	if e.Category != "" {
		out["category"] = e.Category
	}
	if e.Note != "" {
		out["note"] = e.Note
	}
	if e.Counterparty != "" {
		out["counterparty"] = e.Counterparty
	}
	e.Data = out
	return out
}

// ForChain serializes entry for appendEvent.
func (e *EntryPayload) ForChain(schema EntrySchema) EntryPayload {
	data := e.NormalizeData()
	schema = ResolveEntrySchema(schema)
	sid := schema.TemplateID
	if sid == "" {
		sid = TemplateDefault
	}
	return EntryPayload{SchemaID: sid, Data: data}
}

const (
	EventLedgerCreated  = "LedgerCreated"
	EventEntryAdded     = "EntryAdded"
	EventImportBatch    = "ImportBatch"
	EventBatchSealed       = "BatchSealed"
	EventBackupAnchored    = "BackupAnchored"
	EventExternalAnchored  = "ExternalAnchored"
	EventLedgerUpdated     = "LedgerUpdated"
	EventLedgerArchived    = "LedgerArchived"
)
// Collaboration events: EntryProposed, EntryApproved, EntryRejected,
// MemberInvited, MemberJoined, GroupKeyRotated — see collaboration.go
