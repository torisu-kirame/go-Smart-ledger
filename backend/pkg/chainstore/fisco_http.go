package chainstore

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const fiscoBlockLimitDelta int64 = 600

type fiscoHTTPClient struct {
	cfg        FISCOConfig
	httpClient *http.Client
}

func newFiscoHTTPClient(cfg FISCOConfig) *fiscoHTTPClient {
	return &fiscoHTTPClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

type fiscoCallResult struct {
	BlockNumber int    `json:"blockNumber"`
	Output      string `json:"output"`
	Status      int    `json:"status"`
}

type fiscoReceipt struct {
	Status          int    `json:"status"`
	TransactionHash string `json:"transactionHash"`
	BlockNumber     string `json:"blockNumber"`
	Message         string `json:"message"`
}

func (c *fiscoHTTPClient) rpc(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      1,
		"params":  params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.JSONRPCURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, fmt.Errorf("fisco rpc %s (%d): %s", method, out.Error.Code, out.Error.Message)
	}
	return out.Result, nil
}

func (c *fiscoHTTPClient) getBlockNumber(ctx context.Context) (uint64, error) {
	raw, err := c.rpc(ctx, "getBlockNumber", nil)
	if err != nil {
		return 0, err
	}
	return parseFiscoBlockNumber(raw), nil
}

func (c *fiscoHTTPClient) callContract(ctx context.Context, contract common.Address, data []byte) ([]byte, error) {
	callArg := map[string]interface{}{
		"to":   strings.ToLower(contract.Hex()[2:]),
		"data": hex.EncodeToString(data),
	}
	raw, err := c.rpc(ctx, "call", []interface{}{callArg})
	if err != nil {
		return nil, err
	}
	var cr fiscoCallResult
	if err := json.Unmarshal(raw, &cr); err != nil {
		return nil, err
	}
	if cr.Status != 0 {
		return nil, fmt.Errorf("fisco call status %d", cr.Status)
	}
	if cr.Output == "" || cr.Output == "0x" {
		return nil, nil
	}
	return common.FromHex(cr.Output), nil
}

func (c *fiscoHTTPClient) sendSignedTransaction(ctx context.Context, signedHex string) (*fiscoReceipt, error) {
	if !strings.HasPrefix(signedHex, "0x") {
		signedHex = "0x" + signedHex
	}
	withProof := false
	raw, err := c.rpc(ctx, "sendTransaction", []interface{}{
		c.cfg.GroupID,
		c.cfg.NodeName,
		signedHex,
		withProof,
	})
	if err != nil {
		return nil, err
	}
	var receipt fiscoReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		// some nodes return tx hash string directly
		var hash string
		if json.Unmarshal(raw, &hash) == nil && hash != "" {
			receipt.TransactionHash = hash
			receipt.Status = 0
			return &receipt, nil
		}
		return nil, err
	}
	if receipt.Status != 0 {
		msg := receipt.Message
		if msg == "" {
			msg = fmt.Sprintf("status %d", receipt.Status)
		}
		return nil, fmt.Errorf("fisco: transaction failed: %s", msg)
	}
	return &receipt, nil
}

func (c *fiscoHTTPClient) getTransactionReceipt(ctx context.Context, txHash string) (*fiscoReceipt, error) {
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}
	raw, err := c.rpc(ctx, "getTransactionReceipt", []interface{}{
		c.cfg.GroupID,
		c.cfg.NodeName,
		txHash,
		false,
	})
	if err != nil {
		return nil, err
	}
	var receipt fiscoReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return nil, err
	}
	return &receipt, nil
}

func parseFiscoBlockNumber(raw json.RawMessage) uint64 {
	var hexStr string
	if json.Unmarshal(raw, &hexStr) == nil && len(hexStr) > 2 && hexStr[:2] == "0x" {
		var n uint64
		_, _ = fmt.Sscanf(hexStr, "0x%x", &n)
		return n
	}
	var n uint64
	_ = json.Unmarshal(raw, &n)
	return n
}
