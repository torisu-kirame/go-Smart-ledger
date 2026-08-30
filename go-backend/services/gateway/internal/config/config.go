package config

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/registry"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string `json:",default=smart-ledger-access-secret-change-me"`
	} `json:"Auth"`
	Upstreams struct {
		Auth    string `json:",default=http://127.0.0.1:28887"`
		Ledger  string `json:",default=http://127.0.0.1:28888"`
		Storage string `json:",default=http://127.0.0.1:28890"`
	} `json:"Upstreams"`
	OpenClaw struct {
		GatewayURL   string `json:",default=http://127.0.0.1:18789"`
		GatewayToken string `json:",optional"`
		AgentModel   string `json:",default=openclaw/default"`
	} `json:"OpenClaw"`
	Agent struct {
		ConfigPath string `json:",default=./data/agent/config"`
	} `json:"Agent"`
	Discovery struct {
		Etcd registry.EtcdConfig `json:"Etcd"`
	} `json:"Discovery"`
	Cors struct {
		AllowedOrigins []string `json:",default=http://localhost:25173"`
	} `json:"Cors"`
	Security SecurityConfig `json:"Security"`
}

// SecurityConfig groups production hardening options (F27).
type SecurityConfig struct {
	TrustForwardedProto bool `json:",default=true"`
	RateLimit           struct {
		Enabled   bool `json:",default=true"`
		GlobalRPM int  `json:",default=600"`
		Burst     int  `json:",default=80"`
		AuthRPM   int  `json:",default=30"`
		AuthBurst int  `json:",default=10"`
	} `json:"RateLimit"`
}
