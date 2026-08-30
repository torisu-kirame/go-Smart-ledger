package mount

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/gateway/internal/ai"
)

// AIOptions configures the embedded OpenClaw agent.
type AIOptions struct {
	OpenClawURL   string
	OpenClawToken string
	AgentModel    string
	LedgerURL     string
	AgentRoot     string
}

// RegisterAI mounts /api/v1/ai/* handlers.
func RegisterAI(r router.Registrar, opts AIOptions) {
	if strings.TrimSpace(opts.LedgerURL) == "" {
		opts.LedgerURL = "http://127.0.0.1:28080"
	}
	h := ai.NewHandlers(ai.Options{
		OpenClawURL:   opts.OpenClawURL,
		OpenClawToken: opts.OpenClawToken,
		AgentModel:    opts.AgentModel,
		LedgerURL:     opts.LedgerURL,
		AgentRoot:     opts.AgentRoot,
	})
	r.Add(http.MethodPost, "/api/v1/ai/chat", h.Chat)
	r.Add(http.MethodPost, "/api/v1/ai/test", h.Test)
	r.Add(http.MethodPost, "/api/v1/ai/agent/load", h.Load)
	r.Add(http.MethodPost, "/api/v1/ai/agent/save", h.Save)
}

// SelfURL builds loopback base URL for in-process ledger tool calls.
func SelfURL(port int) string {
	if port <= 0 {
		port = 28080
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}
