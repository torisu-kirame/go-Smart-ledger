package chainstore

import "testing"

func TestFISCOConfigDefaults(t *testing.T) {
	cfg := FISCOConfig{}.withDefaults()
	if cfg.GroupID != "group0" {
		t.Fatalf("GroupID = %q", cfg.GroupID)
	}
	if cfg.JSONRPCURL != "http://127.0.0.1:20200" {
		t.Fatalf("JSONRPCURL = %q", cfg.JSONRPCURL)
	}
}

func TestParseRPCHostPort(t *testing.T) {
	host, port, err := parseHostPortForTest("http://127.0.0.1:20200")
	if err != nil || host != "127.0.0.1" || port != 20200 {
		t.Fatalf("host=%q port=%d err=%v", host, port, err)
	}
	host, port, err = parseHostPortForTest("http://host.docker.internal:8545")
	if err != nil || host != "host.docker.internal" || port != 8545 {
		t.Fatalf("host=%q port=%d err=%v", host, port, err)
	}
}
