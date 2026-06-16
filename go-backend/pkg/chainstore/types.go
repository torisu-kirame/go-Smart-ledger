package chainstore

import "encoding/json"

// Backend identifies the authoritative ledger chain implementation.
type Backend string

const (
	BackendMiniLedger Backend = "miniledger"
	BackendFISCO      Backend = "fisco"
)

// TxRequest is a generic key-value chain write (MiniLedger KV or FISCO table row).
type TxRequest struct {
	Key     string          `json:"key,omitempty"`
	Value   json.RawMessage `json:"value,omitempty"`
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// StateRow is one queried state entry.
type StateRow struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// Status summarizes node / block height.
type Status struct {
	Height      uint64 `json:"height"`
	Uptime      string `json:"uptime,omitempty"`
	Role        string `json:"role,omitempty"`
	Backend     string `json:"backend,omitempty"`
	ExplorerURL string `json:"explorerUrl,omitempty"`
}
