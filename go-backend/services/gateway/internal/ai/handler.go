package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/rest/httpx"
	xerrors "github.com/zeromicro/x/errors"
)

// Options configures the AI handlers (OpenClaw-backed).
type Options struct {
	OpenClawURL   string
	OpenClawToken string
	AgentModel    string
	LedgerURL     string
	AgentRoot     string
}

// Handlers serves /api/v1/ai/* using OpenClaw as the agent core.
type Handlers struct {
	opts Options
}

func NewHandlers(opts Options) *Handlers {
	if opts.AgentModel == "" {
		opts.AgentModel = "openclaw/default"
	}
	if opts.OpenClawURL == "" {
		opts.OpenClawURL = "http://openclaw-gateway:18789"
	}
	if opts.LedgerURL == "" {
		opts.LedgerURL = "http://ledger-api:28888"
	}
	if opts.AgentRoot == "" {
		opts.AgentRoot = "/data/agent/config"
	}
	return &Handlers{opts: opts}
}

func (h *Handlers) client(gatewayURL, gatewayToken string) *OpenClawClient {
	url := strings.TrimSpace(gatewayURL)
	if url == "" {
		url = h.opts.OpenClawURL
	}
	tok := strings.TrimSpace(gatewayToken)
	if tok == "" {
		tok = h.opts.OpenClawToken
	}
	return &OpenClawClient{BaseURL: url, Token: tok, Timeout: 180 * time.Second}
}

type chatReq struct {
	BaseURL       string        `json:"baseUrl"`
	APIKey        string        `json:"apiKey"`
	Model         string        `json:"model"`
	GatewayURL    string        `json:"gatewayUrl"`
	GatewayToken  string        `json:"gatewayToken"`
	Messages      []ChatMessage `json:"messages"`
	Stream        bool          `json:"stream"`
	UseTools      bool          `json:"useTools"`
	BoundLedgerID string        `json:"boundLedgerId"`
}

type testReq struct {
	GatewayURL   string `json:"gatewayUrl"`
	GatewayToken string `json:"gatewayToken"`
	Model        string `json:"model"`
}

type storagePaths struct {
	AgentPath       string `json:"agentPath"`
	ChatHistoryPath string `json:"chatHistoryPath"`
}

type storageLoadReq struct {
	storagePaths
	LoadWorkspace bool `json:"loadWorkspace"`
}

type storageSaveReq struct {
	storagePaths
	Messages       []ChatMessage     `json:"messages"`
	WorkspaceFiles map[string]string `json:"workspaceFiles"`
}

// Chat handles POST /api/v1/ai/chat
func (h *Handlers) Chat(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimSpace(r.Header.Get("X-User-Id"))
	if uid == "" {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
		return
	}
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
		return
	}
	if len(req.Messages) == 0 {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "messages required"))
		return
	}
	oc := h.client(req.GatewayURL, req.GatewayToken)
	ocModel := resolveOpenClawModel(req.Model)
	sessionKey := "sl-" + uuid.NewString()

	if req.UseTools {
		userInput := buildUserInput(req.Messages)
		if userInput == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "no user message for agent"))
			return
		}
		// SourceText = current turn only (not chat history), so auto-import / MD prefer
		// do not re-fire on follow-ups that still contain an older table in context.
		currentTurn := latestUserContent(req.Messages)
		if currentTurn == "" {
			currentTurn = userInput
		}
		prefix := buildSystemPrefix(req.Messages, req.BoundLedgerID)
		auth := r.Header.Get("Authorization")
		ledger := &LedgerHTTP{BaseURL: h.opts.LedgerURL, AuthHeader: auth, UserID: uid, SourceText: currentTurn}
		content, err := runToolAgent(
			r.Context(), oc, ledger,
			buildAgentMessages(prefix, userInput),
			req.BoundLedgerID, ocModel, h.opts.AgentModel, sessionKey,
		)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(502, err.Error()))
			return
		}
		if req.Stream {
			writeSSEText(w, content)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"choices": []any{
				map[string]any{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": content,
					},
					"finish_reason": "stop",
				},
			},
		})
		return
	}

	msgs := make([]map[string]any, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		c := strings.TrimSpace(m.Content)
		if c == "" {
			continue
		}
		if role == "" {
			role = "user"
		}
		msgs = append(msgs, map[string]any{"role": role, "content": c})
	}
	body := map[string]any{
		"model":    h.opts.AgentModel,
		"messages": msgs,
		"stream":   req.Stream,
	}

	if req.Stream {
		if err := oc.ChatCompletionStream(r.Context(), body, ocModel, sessionKey, w); err != nil {
			if w.Header().Get("Content-Type") == "" {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(502, err.Error()))
			}
		}
		return
	}
	resp, err := oc.ChatCompletion(r.Context(), body, ocModel, sessionKey)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(502, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

// Test handles POST /api/v1/ai/test
func (h *Handlers) Test(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.Header.Get("X-User-Id")) == "" {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
		return
	}
	var req testReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	oc := h.client(req.GatewayURL, req.GatewayToken)
	_, err := oc.ChatCompletion(r.Context(), map[string]any{
		"model":    h.opts.AgentModel,
		"messages": []map[string]any{{"role": "user", "content": "ping"}},
		"stream":   false,
	}, resolveOpenClawModel(req.Model), "sl-test-"+uuid.NewString())
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(502, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, map[string]bool{"ok": true})
}

// Load handles POST /api/v1/ai/agent/load
func (h *Handlers) Load(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.Header.Get("X-User-Id")) == "" {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
		return
	}
	var req storageLoadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
		return
	}
	st := &Storage{ConfigRoot: h.opts.AgentRoot}
	msgs, ws, updated, err := st.Load(req.AgentPath, req.ChatHistoryPath, req.LoadWorkspace)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, map[string]any{
		"messages":       msgs,
		"workspaceFiles": ws,
		"updatedAt":      updated,
	})
}

// Save handles POST /api/v1/ai/agent/save
func (h *Handlers) Save(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.Header.Get("X-User-Id")) == "" {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
		return
	}
	var req storageSaveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
		return
	}
	st := &Storage{ConfigRoot: h.opts.AgentRoot}
	if err := st.Save(req.AgentPath, req.ChatHistoryPath, req.Messages, req.WorkspaceFiles); err != nil {
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, map[string]bool{"ok": true})
}

func writeSSEText(w http.ResponseWriter, content string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.OkJson(w, map[string]any{
			"choices": []any{map[string]any{
				"message": map[string]any{"role": "assistant", "content": content},
			}},
		})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	// Chunk by runes so multi-byte UTF-8 (e.g. Chinese) is never split mid-character.
	runes := []rune(content)
	const runeChunk = 24
	for i := 0; i < len(runes); i += runeChunk {
		end := i + runeChunk
		if end > len(runes) {
			end = len(runes)
		}
		part := string(runes[i:end])
		payload, _ := json.Marshal(map[string]any{
			"choices": []any{map[string]any{
				"index": 0,
				"delta": map[string]any{"content": part},
			}},
		})
		_, _ = io.WriteString(w, "data: "+string(payload)+"\n\n")
		flusher.Flush()
	}
	_, _ = io.WriteString(w, "data: [DONE]\n\n")
	flusher.Flush()
}
