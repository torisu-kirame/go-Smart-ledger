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

func (c FISCOConfig) withDefaults() FISCOConfig {
	if strings.TrimSpace(c.GroupID) == "" {
		c.GroupID = "group0"
	}
	if strings.TrimSpace(c.ChainID) == "" {
		c.ChainID = "chain0"
	}
	if strings.TrimSpace(c.JSONRPCURL) == "" {
		c.JSONRPCURL = "http://127.0.0.1:20200"
	}
	return c
}

func loadFISCOConfig(cfg FISCOConfig) (FISCOConfig, error) {
	cfg = cfg.withDefaults()
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_JSONRPC")); v != "" {
		cfg.JSONRPCURL = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_GROUP_ID")); v != "" {
		cfg.GroupID = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_CHAIN_ID")); v != "" {
		cfg.ChainID = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_REGISTRY_CONTRACT")); v != "" {
		cfg.RegistryContract = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_FISCO_PRIVATE_KEY")); v != "" {
		cfg.PrivateKeyHex = v
	}
	if cfg.PrivateKeyHex == "" && cfg.PrivateKeyPath != "" {
		raw, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return cfg, fmt.Errorf("fisco: read PrivateKeyPath: %w", err)
		}
		cfg.PrivateKeyHex = strings.TrimSpace(string(raw))
	}
	return cfg, nil
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

func parseHostPort(rawURL string) (host string, port int, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", 0, err
	}
	host = u.Hostname()
	if host == "" {
		return "", 0, fmt.Errorf("fisco: invalid JSONRPCURL %q", rawURL)
	}
	portStr := u.Port()
	if portStr == "" {
		if u.Scheme == "https" {
			port = 443
		} else {
			port = 20200
		}
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, err
		}
	}
	return host, port, nil
}

func parseHostPortForTest(rawURL string) (string, int, error) {
	return parseHostPort(rawURL)
}
