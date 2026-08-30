package chainstore

import "context"

// Store is the authoritative ledger chain backend (MiniLedger).
type Store interface {
	Backend() Backend
	Ping(ctx context.Context) error
	Status(ctx context.Context) (*Status, error)
	Submit(ctx context.Context, tx TxRequest) error
	Query(ctx context.Context, sql string, params ...interface{}) ([]StateRow, error)
	// GetRaw proxies explorer-style GET paths on MiniLedger.
	GetRaw(ctx context.Context, path string) ([]byte, error)
	BaseURL() string
}
