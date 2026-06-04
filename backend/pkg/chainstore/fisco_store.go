package chainstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FISCOStore implements Store against FISCO BCOS (JSON-RPC + LedgerRegistry contract).
// v0.14.2: RPC ping + status; Submit/Query via contract (see docs/fisco-bcos-migration.md).
type FISCOStore struct {
	cfg        FISCOConfig
	httpClient *http.Client
}

func NewFISCO(cfg FISCOConfig) (*FISCOStore, error) {
	if cfg.JSONRPCURL == "" {
		return nil, fmt.Errorf("fisco: JSONRPCURL required")
	}
	return &FISCOStore{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *FISCOStore) Backend() Backend { return BackendFISCO }

func (s *FISCOStore) Ping(ctx context.Context) error {
	_, err := s.fiscoRPC(ctx, "getBlockNumber", []interface{}{s.cfg.GroupID})
	return err
}

func (s *FISCOStore) Status(ctx context.Context) (*Status, error) {
	raw, err := s.fiscoRPC(ctx, "getBlockNumber", []interface{}{s.cfg.GroupID})
	if err != nil {
		return nil, err
	}
	return &Status{
		Height:      parseFiscoBlockNumber(raw),
		Backend:     string(BackendFISCO),
		ExplorerURL: s.cfg.ExplorerURL,
		Role:        "fisco-bcos",
	}, nil
}

func (s *FISCOStore) Submit(ctx context.Context, tx TxRequest) error {
	if s.cfg.RegistryContract == "" {
		return fmt.Errorf("fisco: RegistryContract not deployed (see backend/contracts/fisco and docs/fisco-bcos-migration.md)")
	}
	_ = ctx
	_ = tx
	return fmt.Errorf("fisco: Submit not implemented yet (migration v0.15.0-fisco.3)")
}

func (s *FISCOStore) Query(ctx context.Context, sql string, params ...interface{}) ([]StateRow, error) {
	_ = ctx
	_ = sql
	_ = params
	return nil, fmt.Errorf("fisco: Query not implemented yet (migration v0.15+)")
}

func (s *FISCOStore) GetRaw(ctx context.Context, path string) ([]byte, error) {
	st, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"backend": "fisco",
		"path":    path,
		"height":  st.Height,
		"hint":    "Use FISCO block explorer or WeBASE",
	})
}

func (s *FISCOStore) BaseURL() string { return s.cfg.JSONRPCURL }

func (s *FISCOStore) fiscoRPC(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      1,
		"params":  params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.JSONRPCURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, fmt.Errorf("fisco rpc %s: %s", method, out.Error.Message)
	}
	return out.Result, nil
}

func parseFiscoBlockNumber(raw json.RawMessage) uint64 {
	var hex string
	if json.Unmarshal(raw, &hex) == nil && len(hex) > 2 && hex[:2] == "0x" {
		var n uint64
		_, _ = fmt.Sscanf(hex, "0x%x", &n)
		return n
	}
	var n uint64
	_ = json.Unmarshal(raw, &n)
	return n
}
