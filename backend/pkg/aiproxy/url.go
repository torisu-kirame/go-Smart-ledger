package aiproxy

import (
	"net"
	"net/url"
	"strings"
)

// normalizeLLMBaseURL validates outbound LLM API base URLs (SSRF-safe).
// HTTPS: allowed for public hosts. HTTP: only local/docker allowlist (Ollama etc.).
func normalizeLLMBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidBaseURL
	}
	host := strings.ToLower(u.Hostname())
	switch u.Scheme {
	case "https":
		if isBlockedHost(host) {
			return "", ErrInvalidBaseURL
		}
	case "http":
		if !isAllowedLocalLLMHost(host) {
			return "", ErrInvalidBaseURL
		}
	default:
		return "", ErrInvalidBaseURL
	}
	return ensureOpenAIV1Suffix(raw), nil
}

func ensureOpenAIV1Suffix(raw string) string {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "/")
	if strings.HasSuffix(raw, "/v1") {
		return raw
	}
	return raw + "/v1"
}

func isAllowedLocalLLMHost(host string) bool {
	allowed := map[string]bool{
		"127.0.0.1":            true,
		"localhost":            true,
		"::1":                  true,
		"ollama":               true,
		"host.docker.internal": true,
	}
	return allowed[host]
}

func isBlockedHost(host string) bool {
	if host == "" {
		return true
	}
	if isAllowedLocalLLMHost(host) {
		// Cloud keys should not target loopback over HTTPS in production proxy.
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}
