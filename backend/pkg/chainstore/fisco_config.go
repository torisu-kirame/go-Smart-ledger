package chainstore

import (
	"crypto/ecdsa"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func loadFISCOConfig(cfg FISCOConfig) (FISCOConfig, error) {
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_PRIVATE_KEY")); v != "" {
		cfg.PrivateKeyHex = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_JSONRPC")); v != "" {
		cfg.JSONRPCURL = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_GROUP_ID")); v != "" {
		cfg.GroupID = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_REGISTRY")); v != "" {
		cfg.RegistryContract = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_CHAIN_ID")); v != "" {
		cfg.ChainID = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_DISABLE_SSL")); v != "" {
		cfg.DisableSsl = strings.EqualFold(v, "1") || strings.EqualFold(v, "true")
	}
	if cfg.PrivateKeyHex == "" && cfg.PrivateKeyPath != "" {
		raw, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return cfg, fmt.Errorf("fisco: read PrivateKeyPath: %w", err)
		}
		cfg.PrivateKeyHex = strings.TrimSpace(string(raw))
	}
	return cfg.withDefaults(), nil
}

func (c FISCOConfig) withDefaults() FISCOConfig {
	if c.GroupID == "" {
		c.GroupID = "group0"
	}
	if c.JSONRPCURL == "" {
		c.JSONRPCURL = "http://127.0.0.1:20200"
	}
	return c
}

func (c FISCOConfig) rpcHostPort() (string, int, error) {
	u, err := url.Parse(c.JSONRPCURL)
	if err != nil {
		return "", 0, fmt.Errorf("fisco: invalid JSONRPCURL: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return "", 0, fmt.Errorf("fisco: JSONRPCURL missing host")
	}
	port := 20200
	if p := u.Port(); p != "" {
		port, err = strconv.Atoi(p)
		if err != nil {
			return "", 0, fmt.Errorf("fisco: invalid port in JSONRPCURL: %w", err)
		}
	}
	return host, port, nil
}

func (c FISCOConfig) registryAddr() (common.Address, error) {
	if c.RegistryContract == "" {
		return common.Address{}, fmt.Errorf("fisco: RegistryContract not configured")
	}
	if !common.IsHexAddress(c.RegistryContract) {
		return common.Address{}, fmt.Errorf("fisco: invalid RegistryContract %q", c.RegistryContract)
	}
	return common.HexToAddress(c.RegistryContract), nil
}

func (c FISCOConfig) privateKey() (*ecdsa.PrivateKey, error) {
	if c.PrivateKeyHex == "" {
		return nil, fmt.Errorf("fisco: PrivateKeyHex or PrivateKeyPath required for Submit")
	}
	hexKey := strings.TrimPrefix(strings.TrimSpace(c.PrivateKeyHex), "0x")
	return crypto.HexToECDSA(hexKey)
}

func parseHostPortForTest(raw string) (string, int, error) {
	cfg := FISCOConfig{JSONRPCURL: raw}.withDefaults()
	return cfg.rpcHostPort()
}
