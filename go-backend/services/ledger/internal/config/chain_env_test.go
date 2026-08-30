package config

import (
	"testing"
)

func TestApplyChainEnv(t *testing.T) {
	t.Setenv("SL_CHAIN_BACKEND", "miniledger")
	c := &Config{}
	c.ApplyChainEnv()
	if c.Chain.Backend != "miniledger" {
		t.Fatalf("backend=%q", c.Chain.Backend)
	}
}

func TestApplyChainEnvEmptyKeepsDefault(t *testing.T) {
	t.Setenv("SL_CHAIN_BACKEND", "")
	c := &Config{}
	c.Chain.Backend = "miniledger"
	c.ApplyChainEnv()
	if c.Chain.Backend != "miniledger" {
		t.Fatalf("backend=%q", c.Chain.Backend)
	}
}
