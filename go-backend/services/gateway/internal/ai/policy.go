package ai

import (
	"fmt"
	"regexp"
	"strings"
)

var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PATCH": true, "PUT": true, "DELETE": true,
}

var allowedPrefixes = []string{
	"/api/v1/ledgers",
	"/api/v1/entry-templates",
	"/api/v1/friends",
	"/api/v1/teams",
	"/api/v1/import",
	"/api/v1/chain",
	"/api/v1/users/me",
	"/api/v1/users/search",
}

var blockedSubstr = []string{
	"/auth/", "/ai/", "/storage/", "/delete-account", "/verify-password",
	"/public-key", "/encryption/", "/restore/commit", "/me/avatar",
}

var placeholderRE = regexp.MustCompile(`(?i)\{(id|ledgerId|ledger_id|tableId|table_id|teamId|team_id)\}`)

func normalizeAPIPath(raw, boundLedgerID string) (string, error) {
	p := strings.TrimSpace(raw)
	if p == "" {
		return "", fmt.Errorf("path required")
	}
	if strings.Contains(p, "://") || strings.Contains(p, "..") || strings.Contains(p, `\`) {
		return "", fmt.Errorf("path must be a relative API path")
	}
	if strings.Contains(p, "?") {
		return "", fmt.Errorf("put query params in the query object, not in path")
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if strings.HasPrefix(p, "/v1/") {
		p = "/api" + p
	}
	if !strings.HasPrefix(p, "/api/v1/") {
		return "", fmt.Errorf("path must start with /api/v1/")
	}
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	lid := strings.TrimSpace(boundLedgerID)
	if lid != "" {
		p = placeholderRE.ReplaceAllStringFunc(p, func(m string) string {
			key := strings.ToLower(strings.Trim(m, "{}"))
			if key == "id" || key == "ledgerid" || key == "ledger_id" {
				return lid
			}
			return m
		})
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p, nil
}

func assertAPIAllowed(method, p string) error {
	m := strings.ToUpper(strings.TrimSpace(method))
	if !allowedMethods[m] {
		return fmt.Errorf("method not allowed: %s", method)
	}
	lower := strings.ToLower(p)
	for _, bad := range blockedSubstr {
		if strings.Contains(lower, bad) {
			return fmt.Errorf("path blocked for AI: contains %s", strings.Trim(bad, "/"))
		}
	}
	ok := false
	for _, pref := range allowedPrefixes {
		if p == pref || strings.HasPrefix(p, pref+"/") {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("path not in AI allowlist")
	}
	return nil
}
