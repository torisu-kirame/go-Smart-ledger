package chainstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type fiscoRegistryClient struct {
	cfg      FISCOConfig
	http     *fiscoHTTPClient
	contract common.Address
	abi      abi.ABI
}

func newFiscoRegistryClient(cfg FISCOConfig) (*fiscoRegistryClient, error) {
	if cfg.RegistryContract == "" {
		return nil, fmt.Errorf("fisco: RegistryContract not configured")
	}
	parsed, err := abi.JSON(strings.NewReader(ledgerRegistryABI))
	if err != nil {
		return nil, err
	}
	contract, err := cfg.registryAddr()
	if err != nil {
		return nil, err
	}
	return &fiscoRegistryClient{
		cfg:      cfg,
		http:     newFiscoHTTPClient(cfg),
		contract: contract,
		abi:      parsed,
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
	priv, err := c.cfg.privateKey()
	if err != nil {
		return err
	}
	data, err := c.abi.Pack("putState", ledgerID, key, value)
	if err != nil {
		return err
	}
	signedHex, err := signFiscoContractCall(ctx, c.http, c.cfg, c.contract, data, priv)
	if err != nil {
		return err
	}
	_, err = c.http.sendSignedTransaction(ctx, signedHex)
	return err
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
