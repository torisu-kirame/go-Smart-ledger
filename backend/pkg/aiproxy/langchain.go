package aiproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// AgentChatRequest is the LangChain-backed chat payload from the console.
type AgentChatRequest struct {
	BaseURL         string        `json:"baseUrl"`
	APIKey          string        `json:"apiKey,omitempty"`
	Model           string        `json:"model"`
	Messages        []ChatMessage `json:"messages"`
	Stream          bool          `json:"stream"`
	UseTools        bool          `json:"useTools,omitempty"`
	BoundLedgerID   string        `json:"boundLedgerId,omitempty"`
}

// ProxyChatContext carries authenticated ledger access for tool calling.
type ProxyChatContext struct {
	UserID string
	Ledger LedgerToolBackend
}

// TestAgentRequest probes LLM connectivity.
type TestAgentRequest struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey,omitempty"`
	Model   string `json:"model"`
}

// TestAgentChat sends a minimal completion to verify provider settings.
func TestAgentChat(ctx context.Context, req TestAgentRequest) error {
	llm, err := newLangChainLLM(req.BaseURL, req.APIKey, req.Model)
	if err != nil {
		return err
	}
	_, err = llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("ping")},
		},
	})
	return err
}

// ProxyAgentChat runs LangChainGo against an OpenAI-compatible LLM endpoint.
func ProxyAgentChat(w http.ResponseWriter, r *http.Request, req AgentChatRequest, pctx *ProxyChatContext) error {
	if req.Model == "" {
		return fmt.Errorf("model required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages required")
	}
	llm, err := newLangChainLLM(req.BaseURL, req.APIKey, req.Model)
	if err != nil {
		return err
	}
	ctx := r.Context()

	if req.UseTools && pctx != nil && pctx.Ledger != nil && strings.TrimSpace(pctx.UserID) != "" {
		return proxyAgentChatWithTools(w, r, ctx, llm, req, pctx)
	}

	msgs := toLangChainMessages(req.Messages)

	if req.Stream {
		flusher, _ := w.(http.Flusher)
		sink := &sseSink{w: w, flusher: flusher}
		resp, err := llm.GenerateContent(ctx, msgs, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			if len(chunk) == 0 {
				return nil
			}
			return sink.emit(string(chunk))
		}))
		if err != nil {
			if !sink.started {
				return err
			}
			_ = sink.emitError(err.Error())
			return nil
		}
		if !sink.started && resp != nil && len(resp.Choices) > 0 {
			if c := strings.TrimSpace(resp.Choices[0].Content); c != "" {
				if err := sink.emit(c); err != nil {
					return err
				}
			}
		}
		if !sink.started {
			return fmt.Errorf("LLM returned empty streaming response")
		}
		return sink.finish()
	}

	resp, err := llm.GenerateContent(ctx, msgs)
	if err != nil {
		return err
	}
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Content
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	})
}

func proxyAgentChatWithTools(w http.ResponseWriter, r *http.Request, ctx context.Context, llm *openai.LLM, req AgentChatRequest, pctx *ProxyChatContext) error {
	input := buildAgentInput(req.Messages)
	if input == "" {
		return fmt.Errorf("no user message for agent")
	}
	prefix := buildAgentSystemPrefix(req.Messages, req.BoundLedgerID)
	content, err := RunLedgerAgent(ctx, llm, pctx.Ledger, pctx.UserID, req.BoundLedgerID, input, prefix)
	if err != nil {
		return err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("agent returned empty response")
	}

	if req.Stream {
		flusher, _ := w.(http.Flusher)
		sink := &sseSink{w: w, flusher: flusher}
		for _, chunk := range chunkText(content, 64) {
			if err := sink.emit(chunk); err != nil {
				return err
			}
		}
		return sink.finish()
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	})
}

func chunkText(s string, size int) []string {
	if size <= 0 || len(s) <= size {
		return []string{s}
	}
	runes := []rune(s)
	out := make([]string, 0, len(runes)/size+1)
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}

func newLangChainLLM(baseURL, apiKey, model string) (*openai.LLM, error) {
	norm, err := normalizeLLMBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	opts := []openai.Option{
		openai.WithBaseURL(norm),
		openai.WithModel(strings.TrimSpace(model)),
	}
	if key := strings.TrimSpace(apiKey); key != "" {
		opts = append(opts, openai.WithToken(key))
	}
	return openai.New(opts...)
}

func toLangChainMessages(in []ChatMessage) []llms.MessageContent {
	out := make([]llms.MessageContent, 0, len(in)+1)
	hasSystem := false
	for _, m := range in {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		switch role {
		case "system":
			hasSystem = true
			out = append(out, llms.MessageContent{
				Role:  llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextPart(content)},
			})
		case "assistant":
			out = append(out, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(content)},
			})
		default:
			out = append(out, llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart(content)},
			})
		}
	}
	if !hasSystem {
		if prompt := DefaultAgentSystemPrompt(); prompt != "" {
			out = append([]llms.MessageContent{{
				Role:  llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextPart(prompt)},
			}}, out...)
		}
	}
	return out
}

// DefaultWorkspaceFiles returns embedded workspace markdown for the agent editor.
func DefaultWorkspaceFiles() map[string]string {
	out := map[string]string{}
	for _, name := range []string{"AGENTS.md", "TOOLS.md"} {
		raw, err := embeddedWorkspace.ReadFile("workspace/" + name)
		if err == nil {
			out[name] = string(raw)
		}
	}
	return out
}

func writeOpenAISSEDelta(w http.ResponseWriter, flusher http.Flusher, delta string) error {
	payload, err := json.Marshal(map[string]any{
		"choices": []map[string]any{
			{
				"index": 0,
				"delta": map[string]string{"content": delta},
			},
		},
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	if flusher != nil {
		flusher.Flush()
	}
	return nil
}

type sseSink struct {
	w       http.ResponseWriter
	flusher http.Flusher
	started bool
}

func (s *sseSink) emit(delta string) error {
	if strings.TrimSpace(delta) == "" {
		return nil
	}
	if !s.started {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("Connection", "keep-alive")
		s.w.WriteHeader(http.StatusOK)
		s.started = true
	}
	return writeOpenAISSEDelta(s.w, s.flusher, delta)
}

func (s *sseSink) emitError(msg string) error {
	if !s.started {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("Connection", "keep-alive")
		s.w.WriteHeader(http.StatusOK)
		s.started = true
	}
	payload, err := json.Marshal(map[string]any{
		"error": map[string]string{"message": msg},
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", payload); err != nil {
		return err
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	_, _ = fmt.Fprint(s.w, "data: [DONE]\n\n")
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}

func (s *sseSink) finish() error {
	if !s.started {
		return fmt.Errorf("no streaming content")
	}
	_, err := fmt.Fprint(s.w, "data: [DONE]\n\n")
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return err
}

// ParseAgentChatRequest accepts LangChain payload or legacy OpenClaw fields.
func ParseAgentChatRequest(raw json.RawMessage) (AgentChatRequest, error) {
	var req AgentChatRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, err
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		var legacy struct {
			GatewayURL    string `json:"gatewayUrl"`
			GatewayToken  string `json:"gatewayToken"`
			OpenClawModel string `json:"openclawModel"`
		}
		if json.Unmarshal(raw, &legacy) == nil && strings.TrimSpace(legacy.GatewayURL) != "" {
			return req, fmt.Errorf("openclaw gateway is deprecated; configure provider baseUrl/apiKey/model in Settings → AI")
		}
	}
	return req, nil
}

// ParseTestAgentRequest accepts LangChain test payload or rejects legacy OpenClaw test.
func ParseTestAgentRequest(raw json.RawMessage) (TestAgentRequest, error) {
	var req TestAgentRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, err
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		var legacy struct {
			GatewayURL string `json:"gatewayUrl"`
		}
		if json.Unmarshal(raw, &legacy) == nil && strings.TrimSpace(legacy.GatewayURL) != "" {
			return req, fmt.Errorf("openclaw gateway is deprecated; use provider baseUrl/apiKey/model")
		}
	}
	return req, nil
}
