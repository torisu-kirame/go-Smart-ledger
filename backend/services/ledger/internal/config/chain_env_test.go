package config

import (
	"os"
	"testing"
)

func TestApplyChainEnv(t *testing.T) {
	t.Setenv("SL_CHAIN_BACKEND", "fisco")
	c := Config{}
	c.ApplyChainEnv()
	if c.Chain.Backend != "fisco" {
		t.Fatalf("backend=%q", c.Chain.Backend)
	}
}

func TestApplyChainEnvEmpty(t *testing.T) {
	os.Unsetenv("SL_CHAIN_BACKEND")
	c := Config{}
	c.Chain.Backend = "miniledger"
	c.ApplyChainEnv()
	if c.Chain.Backend != "miniledger" {
		t.Fatalf("backend=%q", c.Chain.Backend)
	}
}
