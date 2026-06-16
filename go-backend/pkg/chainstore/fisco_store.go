package chainstore

import (
	"context"
	"fmt"
)

// FISCOStore implements Store against FISCO BCOS 3.x (native JSON-RPC + LedgerRegistry putState/getState).
type FISCOStore struct {
	cfg  FISCOConfig
	http *fiscoHTTPClient
	registry *fiscoRegistryClient
}

func NewFISCO(cfg FISCOConfig) (*FISCOStore, error) {
	cfg, err := loadFISCOConfig(cfg)
	if err != nil {
		return nil, err
	}
	st := &FISCOStore{
		cfg:  cfg,
		http: newFiscoHTTPClient(cfg),
	}
	if cfg.RegistryContract != "" {
		rc, err := newFiscoRegistryClient(cfg)
		if err != nil {
			return nil, err
		}
		st.registry = rc
	}
	return st, nil
}

func (s *FISCOStore) Backend() Backend { return BackendFISCO }

func (s *FISCOStore) Ping(ctx context.Context) error {
	_, err := s.http.getBlockNumber(ctx)
	return err
}

func (s *FISCOStore) Status(ctx context.Context) (*Status, error) {
	height, err := s.http.getBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return &Status{
		Height:      height,
		Backend:     string(BackendFISCO),
		ExplorerURL: s.cfg.ExplorerURL,
		Role:        "fisco-bcos-3.x",
	}, nil
}

func (s *FISCOStore) Submit(ctx context.Context, tx TxRequest) error {
	if s.registry == nil {
		return fmt.Errorf("fisco: RegistryContract not deployed (see backend/contracts/fisco and docs/fisco-bcos-migration.md)")
	}
	delete := txValueIsDelete(tx.Value)
	var value []byte
	if !delete {
		value = append([]byte(nil), tx.Value...)
	}
	ledgerID := ledgerIDFromStateKey(tx.Key)
	if err := s.registry.putState(ctx, ledgerID, tx.Key, value); err != nil {
		return err
	}
	return s.syncIndexes(ctx, tx.Key, value, delete)
}

func (s *FISCOStore) Query(ctx context.Context, sql string, params ...interface{}) ([]StateRow, error) {
	q, err := parseFiscoQuery(sql, params)
	if err != nil {
		return nil, err
	}
	return s.runQuery(ctx, q)
}

func (s *FISCOStore) GetRaw(ctx context.Context, path string) ([]byte, error) {
	return s.http.explorerGet(ctx, path)
}

func (s *FISCOStore) BaseURL() string { return s.cfg.JSONRPCURL }

// HTTP returns the underlying FISCO 3.x JSON-RPC client (for tests / diagnostics).
func (s *FISCOStore) HTTP() *fiscoHTTPClient { return s.http }
