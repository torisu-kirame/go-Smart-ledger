package chainstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrPermanent indicates a chain write failure that should not be retried.
var ErrPermanent = errors.New("chain: permanent error")

// PermanentError wraps a non-retryable submission failure.
func PermanentError(msg string) error {
	return fmt.Errorf("%w: %s", ErrPermanent, msg)
}

// IsRetryable reports whether txqueue should keep retrying this error.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrPermanent) {
		return false
	}
	msg := strings.ToLower(err.Error())
	nonRetry := []string{
		"reverted",
		"invalid registrycontract",
		"privatekeyhex",
		"not configured",
		"unauthorized",
		"no ledger",
		"ledger not found",
	}
	for _, s := range nonRetry {
		if strings.Contains(msg, s) {
			return false
		}
	}
	retryable := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporarily unavailable",
		"block limit",
		"over block limit",
		"waiting receipt",
		"wait mined",
		"nonce",
	}
	for _, s := range retryable {
		if strings.Contains(msg, s) {
			return true
		}
	}
	// FISCO RPC transport / unknown: retry by default (matches MiniLedger queue semantics).
	return true
}

type explorerBlock struct {
	Height       uint64 `json:"height"`
	Hash         string `json:"hash,omitempty"`
	TxCount      int    `json:"txCount,omitempty"`
	Timestamp    any    `json:"timestamp,omitempty"`
	Time         string `json:"time,omitempty"`
	Transactions []any  `json:"transactions,omitempty"`
}

func (c *fiscoHTTPClient) explorerGet(ctx context.Context, path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	base := u.Path
	q := u.Query()

	switch {
	case base == "/blocks/latest":
		latest, err := c.getBlockNumber(ctx)
		if err != nil {
			return nil, err
		}
		b, err := c.getBlockByNumber(ctx, latest, false)
		if err != nil {
			return nil, err
		}
		return json.Marshal(b)
	case strings.HasPrefix(base, "/blocks/"):
		hStr := strings.TrimPrefix(base, "/blocks/")
		height, err := strconv.ParseUint(hStr, 10, 64)
		if err != nil {
			return nil, err
		}
		b, err := c.getBlockByNumber(ctx, height, false)
		if err != nil {
			return nil, err
		}
		return json.Marshal(b)
	case base == "/blocks":
		page, _ := strconv.Atoi(q.Get("page"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}
		blocks, err := c.listBlocks(ctx, page, limit)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"blocks": blocks})
	case base == "/consensus":
		return c.getConsensusJSON(ctx)
	case base == "/peers":
		return c.getPeersJSON(ctx)
	case base == "/tx/recent":
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		txs, err := c.recentTransactions(ctx, limit)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"transactions": txs})
	default:
		st, err := c.getBlockNumber(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{
			"backend": "fisco",
			"version": "3.x",
			"path":    path,
			"height":  st,
			"groupID": c.cfg.GroupID,
		})
	}
}

func (c *fiscoHTTPClient) listBlocks(ctx context.Context, page, limit int) ([]explorerBlock, error) {
	latest, err := c.getBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	start := int64(latest) - int64((page-1)*limit)
	end := start - int64(limit) + 1
	if end < 0 {
		end = 0
	}
	out := make([]explorerBlock, 0, limit)
	for h := start; h >= end; h-- {
		b, err := c.getBlockByNumber(ctx, uint64(h), true)
		if err != nil {
			return out, err
		}
		if b != nil {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (c *fiscoHTTPClient) getBlockByNumber(ctx context.Context, height uint64, onlyTxHash bool) (*explorerBlock, error) {
	raw, err := c.rpc(ctx, "getBlockByNumber", []interface{}{
		c.cfg.GroupID,
		c.cfg.NodeName,
		height,
		false,
		onlyTxHash,
	})
	if err != nil {
		return nil, err
	}
	var fb fiscoBlockJSON
	if err := json.Unmarshal(raw, &fb); err != nil {
		return nil, err
	}
	return fb.normalize(), nil
}

type fiscoBlockJSON struct {
	Hash         string          `json:"hash"`
	Number       json.RawMessage `json:"number"`
	Timestamp    json.RawMessage `json:"timestamp"`
	Transactions json.RawMessage `json:"transactions"`
}

func (fb *fiscoBlockJSON) normalize() *explorerBlock {
	height := parseFiscoBlockNumber(fb.Number)
	ts := parseFiscoTimestamp(fb.Timestamp)
	b := &explorerBlock{
		Height:    height,
		Hash:      fb.Hash,
		Timestamp: ts,
	}
	if ts > 0 {
		b.Time = time.Unix(int64(ts), 0).UTC().Format(time.RFC3339)
	}
	if len(fb.Transactions) > 0 {
		var txs []any
		if json.Unmarshal(fb.Transactions, &txs) == nil {
			b.Transactions = txs
			b.TxCount = len(txs)
		} else {
			var hashes []string
			if json.Unmarshal(fb.Transactions, &hashes) == nil {
				b.TxCount = len(hashes)
				for _, h := range hashes {
					b.Transactions = append(b.Transactions, map[string]string{"hash": h})
				}
			}
		}
	}
	return b
}

func parseFiscoTimestamp(raw json.RawMessage) uint64 {
	if len(raw) == 0 {
		return 0
	}
	var n uint64
	if json.Unmarshal(raw, &n) == nil {
		return n
	}
	var hexStr string
	if json.Unmarshal(raw, &hexStr) == nil && strings.HasPrefix(hexStr, "0x") {
		var v uint64
		_, _ = fmt.Sscanf(hexStr, "0x%x", &v)
		return v
	}
	return 0
}

func (c *fiscoHTTPClient) getConsensusJSON(ctx context.Context) ([]byte, error) {
	sealers, _ := c.rpc(ctx, "getSealerList", []interface{}{c.cfg.GroupID, c.cfg.NodeName})
	consensus, _ := c.rpc(ctx, "getConsensusStatus", []interface{}{c.cfg.GroupID, c.cfg.NodeName})
	pbft, _ := c.rpc(ctx, "getPbftView", []interface{}{c.cfg.GroupID, c.cfg.NodeName})
	out := map[string]any{
		"backend": "fisco",
		"groupID": c.cfg.GroupID,
	}
	if len(sealers) > 0 {
		out["sealers"] = json.RawMessage(sealers)
	}
	if len(consensus) > 0 {
		out["consensusStatus"] = json.RawMessage(consensus)
	}
	if len(pbft) > 0 {
		out["pbftView"] = json.RawMessage(pbft)
	}
	return json.Marshal(out)
}

func (c *fiscoHTTPClient) getPeersJSON(ctx context.Context) ([]byte, error) {
	raw, err := c.rpc(ctx, "getGroupPeers", []interface{}{c.cfg.GroupID, c.cfg.NodeName})
	if err != nil {
		// fallback to getPeers
		raw, err = c.rpc(ctx, "getPeers", []interface{}{c.cfg.GroupID})
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"backend": "fisco",
		"groupID": c.cfg.GroupID,
		"peers":   json.RawMessage(raw),
	})
}

func (c *fiscoHTTPClient) recentTransactions(ctx context.Context, limit int) ([]map[string]any, error) {
	latest, err := c.getBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for h := int64(latest); h >= 0 && len(out) < limit; h-- {
		b, err := c.getBlockByNumber(ctx, uint64(h), false)
		if err != nil || b == nil {
			continue
		}
		for _, tx := range b.Transactions {
			if len(out) >= limit {
				break
			}
			switch v := tx.(type) {
			case map[string]any:
				v["blockHeight"] = b.Height
				out = append(out, v)
			case string:
				out = append(out, map[string]any{"hash": v, "blockHeight": b.Height})
			default:
				out = append(out, map[string]any{"raw": v, "blockHeight": b.Height})
			}
		}
	}
	return out, nil
}
