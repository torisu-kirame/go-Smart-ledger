package config

import "github.com/smart-ledger/go-smart-ledger/backend/pkg/secrets"

// ApplyEnv overlays gateway secrets from the environment (F27).
func (c *Config) ApplyEnv() {
	if v := secrets.String("SL_GATEWAY_ACCESS_SECRET", ""); v != "" {
		c.Auth.AccessSecret = v
	} else {
		c.Auth.AccessSecret = secrets.String("SL_ACCESS_SECRET", c.Auth.AccessSecret)
	}
	secrets.MustNotDefault("SL_ACCESS_SECRET", c.Auth.AccessSecret, "smart-ledger-access-secret-change-me")
}
