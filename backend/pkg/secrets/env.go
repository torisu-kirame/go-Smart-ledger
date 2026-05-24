// Package secrets reads production secrets from environment variables (F27).
package secrets

import (
	"os"
	"strings"
)

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func envBool(key string) bool {
	v := strings.ToLower(env(key))
	return v == "1" || v == "true" || v == "yes"
}

// IsProduction is true when SL_ENV is production or prod.
func IsProduction() bool {
	e := strings.ToLower(env("SL_ENV"))
	return e == "production" || e == "prod"
}

// String returns the env value or fallback.
func String(envKey, fallback string) string {
	if v := env(envKey); v != "" {
		return v
	}
	return fallback
}

// MustNotDefault panics in production when the value is still a known dev default.
func MustNotDefault(envKey, current, devDefault string) {
	if !IsProduction() {
		return
	}
	if current == "" || current == devDefault {
		panic(envKey + " must be set when SL_ENV=production")
	}
}

// CookieSecure returns whether refresh cookies should use the Secure flag.
func CookieSecure(configured bool) bool {
	if envBool("SL_COOKIE_SECURE") || IsProduction() {
		return true
	}
	return configured
}

// CookieDomain returns optional cookie Domain attribute.
func CookieDomain(configured string) string {
	if v := env("SL_COOKIE_DOMAIN"); v != "" {
		return v
	}
	return configured
}
