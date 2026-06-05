package chainstore

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	fiscotypes "github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type fiscoRegistryClient struct {
	cfg      FISCOConfig
	http     *fiscoHTTPClient
	contract common.Address
	abi      abi.ABI
	key      *ecdsa.PrivateKey
}

func newFiscoRegistryClient(cfg FISCOConfig) (*fiscoRegistryClient, error) {
	cfg, err := loadFISCOConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.JSONRPCURL == "" {
		return nil, fmt.Errorf("fisco: JSONRPCURL required")
	}
	parsed, err := abi.JSON(strings.NewReader(ledgerRegistryABI))
	if err != nil {
		return nil, err
	}
	contract, err := cfg.registryAddr()
	if err != nil {
		return nil, err
	}
	var key *ecdsa.PrivateKey
	if cfg.PrivateKeyHex != "" || cfg.PrivateKeyPath != "" {
		key, err = cfg.privateKey()
		if err != nil {
			return nil, err
		}
	}
	return &fiscoRegistryClient{
		cfg:      cfg,
		http:     newFiscoHTTPClient(cfg),
		contract: contract,
		abi:      parsed,
		key:      key,
	}, nil
}

func (c *fiscoRegistryClient) getState(ctx context.Context, ledgerID, key string) ([]byte, error) {
	data, err := c.abi.Pack("getState", ledgerID, key)
	if err != nil {
		return nil, err
	}
	out, err := c.http.callContract(ctx, c.contract, data)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	unpacked, err := c.abi.Unpack("getState", out)
	if err != nil {
		return nil, err
	}
	if len(unpacked) == 0 {
		return nil, nil
	}
	b, _ := unpacked[0].([]byte)
	return b, nil
}

func (c *fiscoRegistryClient) putState(ctx context.Context, ledgerID, key string, value []byte) error {
	if c.key == nil {
		return fmt.Errorf("fisco: PrivateKeyHex or PrivateKeyPath required for Submit")
	}
	data, err := c.abi.Pack("putState", ledgerID, key, value)
	if err != nil {
		return err
	}
	blockNum, err := c.http.getBlockNumber(ctx)
	if err != nil {
		return err
	}
	signedHex, err := c.buildSignedTransactionHex(data, blockNum+uint64(fiscoBlockLimitDelta))
	if err != nil {
		return err
	}
	receipt, err := c.http.sendSignedTransaction(ctx, signedHex)
	if err != nil {
		return err
	}
	if receipt == nil || receipt.TransactionHash == "" {
		return nil
	}
	return c.waitReceipt(ctx, receipt.TransactionHash)
}

func (c *fiscoRegistryClient) waitReceipt(ctx context.Context, txHash string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		receipt, err := c.http.getTransactionReceipt(ctx, txHash)
		if err == nil && receipt != nil {
			if receipt.Status == 0 {
				return nil
			}
			return fmt.Errorf("fisco: putState reverted (tx %s)", txHash)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("fisco: timeout waiting receipt for %s", txHash)
}

// buildSignedTransactionHex encodes a FISCO BCOS 3.x contract call for sendTransaction JSON-RPC.
func (c *fiscoRegistryClient) buildSignedTransactionHex(input []byte, blockLimit uint64) (string, error) {
	chainID := c.cfg.ChainID
	if chainID == "" {
		chainID = "chain0"
	}
	nonce, err := randomFiscoNonce()
	if err != nil {
		return "", err
	}
	tx := fiscotypes.NewTransaction(
		c.contract,
		big.NewInt(0),
		0,
		big.NewInt(0),
		int64(blockLimit),
		input,
		nonce,
		chainID,
		c.cfg.GroupID,
		"",
		c.cfg.IsSMCrypto,
	)
	tx.Data.Abi = ledgerRegistryABI
	signer := fiscotypes.NewEIP155Signer(big.NewInt(0))
	signed, err := fiscotypes.SignTx(tx, signer, c.key)
	if err != nil {
		return "", fmt.Errorf("fisco: sign tx: %w", err)
	}
	return "0x" + hex.EncodeToString(signed.Bytes()), nil
}

func randomFiscoNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (c *fiscoRegistryClient) getJSONIndex(ctx context.Context, ledgerID, key string) ([]string, error) {
	raw, err := c.getState(ctx, ledgerID, key)
	if err != nil || len(raw) == 0 {
		return nil, err
	}
	var keys []string
	if err := json.Unmarshal(raw, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *fiscoRegistryClient) putJSONIndex(ctx context.Context, ledgerID, key string, keys []string) error {
	raw, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	return c.putState(ctx, ledgerID, key, raw)
}
