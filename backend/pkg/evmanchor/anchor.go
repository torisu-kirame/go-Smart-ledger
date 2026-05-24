package evmanchor

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Result is stored on the ledger after a successful external anchor.
type Result struct {
	TxHash      string
	ChainID     uint64
	ChainName   string
	ExplorerURL string
	AnchoredAt  time.Time
}

// Anchorer posts Merkle roots to an EVM contract.
type Anchorer interface {
	Enabled() bool
	Anchor(ctx context.Context, ledgerID, merkleRoot string, seqFrom, seqTo uint64) (*Result, error)
}

type noop struct{}

func (noop) Enabled() bool { return false }

func (noop) Anchor(context.Context, string, string, uint64, uint64) (*Result, error) {
	return nil, nil
}

// Noop returns a disabled anchorer.
func Noop() Anchorer { return noop{} }

// Client anchors roots via LedgerAnchor.anchor on an EVM JSON-RPC endpoint.
type Client struct {
	cfg  Config
	eth  *ethclient.Client
	abi  abi.ABI
	from common.Address
	key  *ecdsa.PrivateKey
}

// New builds an anchor client; returns Noop when disabled.
func New(cfg Config) (Anchorer, error) {
	cfg, err := LoadConfigFromEnv(cfg)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return Noop(), nil
	}
	parsed, err := abi.JSON(strings.NewReader(ledgerAnchorABI))
	if err != nil {
		return nil, err
	}
	key, err := cfg.privateKey()
	if err != nil {
		return nil, err
	}
	if _, err := cfg.contractAddr(); err != nil {
		return nil, err
	}
	eth, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("evm rpc dial: %w", err)
	}
	return &Client{
		cfg:  cfg,
		eth:  eth,
		abi:  parsed,
		from: crypto.PubkeyToAddress(key.PublicKey),
		key:  key,
	}, nil
}

func (c *Client) Enabled() bool { return c != nil && c.eth != nil }

func (c *Client) Anchor(ctx context.Context, ledgerID, merkleRoot string, seqFrom, seqTo uint64) (*Result, error) {
	contract, err := c.cfg.contractAddr()
	if err != nil {
		return nil, err
	}
	data, err := c.abi.Pack("anchor", ledgerIDHash(ledgerID), rootBytes32(merkleRoot), seqFrom, seqTo)
	if err != nil {
		return nil, err
	}
	chainID := new(big.Int).SetUint64(c.cfg.ChainID)
	if chainID.Sign() == 0 {
		id, err := c.eth.ChainID(ctx)
		if err != nil {
			return nil, err
		}
		chainID = id
	}
	nonce, err := c.eth.PendingNonceAt(ctx, c.from)
	if err != nil {
		return nil, err
	}
	gasPrice, err := c.eth.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	call := ethereum.CallMsg{From: c.from, To: &contract, Data: data}
	gas, err := c.eth.EstimateGas(ctx, call)
	if err != nil {
		gas = 300_000
	}
	tx := types.NewTransaction(nonce, contract, big.NewInt(0), gas, gasPrice, data)
	signed, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.key)
	if err != nil {
		return nil, err
	}
	if err := c.eth.SendTransaction(ctx, signed); err != nil {
		return nil, fmt.Errorf("evm send tx: %w", err)
	}
	hash := signed.Hash().Hex()
	var explorer string
	if tpl := c.cfg.ExplorerURLTemplate; tpl != "" {
		explorer = fmt.Sprintf(tpl, hash)
	}
	return &Result{
		TxHash:      hash,
		ChainID:     chainID.Uint64(),
		ChainName:   c.cfg.ChainName,
		ExplorerURL: explorer,
		AnchoredAt:  time.Now().UTC(),
	}, nil
}

func ledgerIDHash(ledgerID string) [32]byte {
	return crypto.Keccak256Hash([]byte("smart-ledger:ledger:" + ledgerID))
}

func rootBytes32(hexRoot string) [32]byte {
	hexRoot = strings.TrimPrefix(strings.TrimSpace(hexRoot), "0x")
	raw, err := hex.DecodeString(hexRoot)
	if err != nil || len(raw) == 0 {
		return crypto.Keccak256Hash([]byte(hexRoot))
	}
	if len(raw) >= 32 {
		var out [32]byte
		copy(out[:], raw[len(raw)-32:])
		return out
	}
	var out [32]byte
	copy(out[32-len(raw):], raw)
	return out
}
