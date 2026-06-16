package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter applies a per-key token bucket (typically client IP).
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*rate.Limiter
	r       rate.Limit
	burst   int
	ttl     time.Duration
}

func New(requestsPerMinute, burst int) *Limiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 60
	}
	if burst < 1 {
		burst = requestsPerMinute / 4
		if burst < 1 {
			burst = 1
		}
	}
	return &Limiter{
		entries: make(map[string]*rate.Limiter),
		r:       rate.Limit(float64(requestsPerMinute) / 60.0),
		burst:   burst,
		ttl:     15 * time.Minute,
	}
}

func (l *Limiter) Allow(key string) bool {
	lim := l.limiter(key)
	return lim.Allow()
}

func (l *Limiter) limiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.entries[key]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.burst)
	l.entries[key] = lim
	if len(l.entries) > 10000 {
		l.entries = make(map[string]*rate.Limiter)
	}
	return lim
}

// ClientIP resolves the client address behind reverse proxies when trusted.
func ClientIP(r *http.Request, trustForwarded bool) string {
	if trustForwarded {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
			return xri
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
