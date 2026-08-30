package config

import "github.com/smart-ledger/go-smart-ledger/go-backend/pkg/secrets"

// ApplyEnv overlays gateway secrets from the environment (F27).
func (c *Config) ApplyEnv() {
	if v := secrets.String("SL_GATEWAY_ACCESS_SECRET", ""); v != "" {
		c.Auth.AccessSecret = v
	} else {
		c.Auth.AccessSecret = secrets.String("SL_ACCESS_SECRET", c.Auth.AccessSecret)
	}
	secrets.MustNotDefault("SL_ACCESS_SECRET", c.Auth.AccessSecret, "smart-ledger-access-secret-change-me")

	if v := secrets.String("OPENCLAW_GATEWAY_URL", ""); v != "" {
		c.OpenClaw.GatewayURL = v
	}
	if v := secrets.String("OPENCLAW_GATEWAY_TOKEN", ""); v != "" {
		c.OpenClaw.GatewayToken = v
	}
	if v := secrets.String("OPENCLAW_AGENT_MODEL", ""); v != "" {
		c.OpenClaw.AgentModel = v
	}
	if v := secrets.String("AGENT_CONFIG_PATH", ""); v != "" {
		c.Agent.ConfigPath = v
	}
	if v := secrets.String("LEDGER_API_URL", ""); v != "" {
		c.Upstreams.Ledger = v
	}
}
