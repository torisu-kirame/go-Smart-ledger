package domain

import "time"

type LedgerType string

const (
	LedgerPrivate LedgerType = "private"
	LedgerMulti   LedgerType = "multi"
)

type Member struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Role    string `json:"role,omitempty"`
}

type LedgerMeta struct {
	ID            string     `json:"id"`
	Type          LedgerType `json:"type"`
	Name          string     `json:"name"`
	CreatorID     string     `json:"creatorId"`
	LedgerAddress string     `json:"ledgerAddress,omitempty"`
	Members       []Member   `json:"members"`
	LatestSeq    uint64     `json:"latestSeq"`
	LatestRoot   string     `json:"latestRoot"`
	AnchorStatus string     `json:"anchorStatus"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
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

type EntryPayload struct {
	Date         string            `json:"date"`
	Type         string            `json:"type"`
	Amount       string            `json:"amount"`
	Category     string            `json:"category,omitempty"`
	Note         string            `json:"note,omitempty"`
	Counterparty string            `json:"counterparty,omitempty"`
	Custom       map[string]string `json:"custom,omitempty"`
}

const (
	EventLedgerCreated = "LedgerCreated"
	EventEntryAdded    = "EntryAdded"
	EventImportBatch   = "ImportBatch"
	EventBatchSealed   = "BatchSealed"
)
