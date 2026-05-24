package middleware

import (
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ratelimit"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/config"
)

// RateLimit returns middleware that limits requests per client IP (F27).
func RateLimit(sec config.SecurityConfig) func(http.HandlerFunc) http.HandlerFunc {
	if !sec.RateLimit.Enabled {
		return func(next http.HandlerFunc) http.HandlerFunc { return next }
	}
	global := ratelimit.New(sec.RateLimit.GlobalRPM, sec.RateLimit.Burst)
	auth := ratelimit.New(sec.RateLimit.AuthRPM, sec.RateLimit.AuthBurst)
	trust := sec.TrustForwardedProto

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/health" {
				next(w, r)
				return
			}
			ip := ratelimit.ClientIP(r, trust)
			lim := global
			if isAuthPath(r.URL.Path) {
				lim = auth
			}
			if !lim.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}
			next(w, r)
		}
	}
}

func isAuthPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/auth/login") ||
		strings.HasPrefix(path, "/api/v1/auth/register") ||
		strings.HasPrefix(path, "/api/v1/auth/captcha")
}
