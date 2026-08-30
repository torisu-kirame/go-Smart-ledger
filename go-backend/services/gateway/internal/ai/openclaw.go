package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenClawClient talks to OpenClaw Gateway OpenAI-compatible HTTP API.
type OpenClawClient struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

func (c *OpenClawClient) base() string {
	u := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if u == "" {
		return ""
	}
	if strings.HasSuffix(u, "/v1") {
		return u
	}
	return u + "/v1"
}

func (c *OpenClawClient) headers(openclawModel, sessionKey string) http.Header {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	tok := strings.TrimSpace(c.Token)
	if tok != "" {
		h.Set("Authorization", "Bearer "+tok)
	}
	if m := strings.TrimSpace(openclawModel); m != "" {
		h.Set("x-openclaw-model", m)
	}
	// Explicit per-turn session avoids sticky/corrupt agent transcripts.
	if sk := strings.TrimSpace(sessionKey); sk != "" {
		h.Set("x-openclaw-session-key", sk)
	}
	return h
}

func (c *OpenClawClient) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 180 * time.Second
}

// ChatCompletion non-streaming.
func (c *OpenClawClient) ChatCompletion(ctx context.Context, body map[string]any, openclawModel, sessionKey string) (map[string]any, error) {
	base := c.base()
	if base == "" {
		return nil, fmt.Errorf("OpenClaw Gateway 未配置：请设置 OPENCLAW_GATEWAY_URL")
	}
	// Never reuse sticky OpenClaw sessions via OpenAI "user" — that caused
	// transcript overflow + EPERM compaction failures on bind-mounted volumes.
	delete(body, "user")
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header = c.headers(openclawModel, sessionKey)
	client := &http.Client{Timeout: c.timeout()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openclaw: %s", formatOpenClawErr(data))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChatCompletionStream streams SSE content deltas to w.
func (c *OpenClawClient) ChatCompletionStream(ctx context.Context, body map[string]any, openclawModel, sessionKey string, w http.ResponseWriter) error {
	base := c.base()
	if base == "" {
		return fmt.Errorf("OpenClaw Gateway 未配置：请设置 OPENCLAW_GATEWAY_URL")
	}
	body["stream"] = true
	delete(body, "user")
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header = c.headers(openclawModel, sessionKey)
	client := &http.Client{Timeout: c.timeout()}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openclaw: %s", formatOpenClawErr(data))
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return werr
			}
			flusher.Flush()
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func formatOpenClawErr(data []byte) string {
	s := strings.TrimSpace(string(data))
	if s == "" {
		return "empty error"
	}
	var obj map[string]any
	if json.Unmarshal(data, &obj) == nil {
		if errObj, ok := obj["error"].(map[string]any); ok {
			msg, _ := errObj["message"].(string)
			typ, _ := errObj["type"].(string)
			if msg != "" && msg != "internal error" {
				if typ != "" {
					return typ + ": " + msg
				}
				return msg
			}
			if cause, ok := errObj["cause"].(string); ok && cause != "" {
				return cause
			}
		}
	}
	return truncate(s, 800)
}

func messageContent(resp map[string]any) string {
	choices, _ := resp["choices"].([]any)
	if len(choices) == 0 {
		return ""
	}
	ch, _ := choices[0].(map[string]any)
	msg, _ := ch["message"].(map[string]any)
	if msg == nil {
		return ""
	}
	switch v := msg["content"].(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		var b strings.Builder
		for _, p := range v {
			m, _ := p.(map[string]any)
			if t, ok := m["text"].(string); ok {
				b.WriteString(t)
			}
		}
		return strings.TrimSpace(b.String())
	default:
		return ""
	}
}

func messageToolCalls(resp map[string]any) []map[string]any {
	choices, _ := resp["choices"].([]any)
	if len(choices) == 0 {
		return nil
	}
	ch, _ := choices[0].(map[string]any)
	msg, _ := ch["message"].(map[string]any)
	if msg == nil {
		return nil
	}
	raw, ok := msg["tool_calls"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, t := range raw {
		if m, ok := t.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
