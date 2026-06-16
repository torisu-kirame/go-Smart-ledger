package config

import "github.com/smart-ledger/go-smart-ledger/go-backend/pkg/secrets"

// ApplyEnv overlays secrets and production flags from the environment (F27).
func (c *Config) ApplyEnv() {
	c.Auth.AccessSecret = secrets.String("SL_ACCESS_SECRET", c.Auth.AccessSecret)
	c.Auth.RefreshSecret = secrets.String("SL_REFRESH_SECRET", c.Auth.RefreshSecret)
	c.Auth.CookieSecure = secrets.CookieSecure(c.Auth.CookieSecure)
	c.Auth.CookieDomain = secrets.CookieDomain(c.Auth.CookieDomain)
	secrets.MustNotDefault("SL_ACCESS_SECRET", c.Auth.AccessSecret, "smart-ledger-access-secret-change-me")
	secrets.MustNotDefault("SL_REFRESH_SECRET", c.Auth.RefreshSecret, "smart-ledger-refresh-secret-change-me")
}
