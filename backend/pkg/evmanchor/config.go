package evmanchor

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Config drives optional EVM/L2 anchoring (F29).
type Config struct {
	Enabled             bool
	RPCURL              string
	ChainID             uint64
	ChainName           string
	Contract            string
	PrivateKeyHex       string
	ExplorerURLTemplate string
}

// LoadConfigFromEnv merges YAML config with SL_EVM_* environment variables.
func LoadConfigFromEnv(base Config) (Config, error) {
	if v := strings.TrimSpace(os.Getenv("SL_EVM_RPC_URL")); v != "" {
		base.RPCURL = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_EVM_CHAIN_ID")); v != "" {
		var id uint64
		if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
			base.ChainID = id
		}
	}
	if v := strings.TrimSpace(os.Getenv("SL_EVM_CONTRACT")); v != "" {
		base.Contract = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_EVM_ANCHOR_PRIVATE_KEY")); v != "" {
		base.PrivateKeyHex = v
	}
	if strings.EqualFold(os.Getenv("SL_EVM_ANCHOR_ENABLED"), "true") ||
		strings.EqualFold(os.Getenv("SL_EVM_ANCHOR_ENABLED"), "1") {
		base.Enabled = true
	}
	return base, nil
}

func (c Config) privateKey() (*ecdsa.PrivateKey, error) {
	if c.PrivateKeyHex == "" {
		return nil, fmt.Errorf("evm anchor private key not configured")
	}
	hexKey := strings.TrimPrefix(strings.TrimSpace(c.PrivateKeyHex), "0x")
	return crypto.HexToECDSA(hexKey)
}

func (c Config) contractAddr() (common.Address, error) {
	if c.Contract == "" {
		return common.Address{}, fmt.Errorf("evm anchor contract address not configured")
	}
	if !common.IsHexAddress(c.Contract) {
		return common.Address{}, fmt.Errorf("invalid contract address %q", c.Contract)
	}
	return common.HexToAddress(c.Contract), nil
}
