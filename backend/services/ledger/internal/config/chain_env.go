package config

import (
	"os"
	"strings"
)

// ApplyChainEnv merges Chain.* from SL_CHAIN_BACKEND and related env vars.
func (c *Config) ApplyChainEnv() {
	if v := strings.TrimSpace(os.Getenv("SL_CHAIN_BACKEND")); v != "" {
		c.Chain.Backend = v
	}
}
