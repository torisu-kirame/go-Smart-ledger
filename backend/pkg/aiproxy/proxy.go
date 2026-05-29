package aiproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidBaseURL = errors.New("ai base url not allowed")
	ErrUpstream       = errors.New("ai upstream error")
)

// ChatRequest mirrors OpenAI chat completions.
type ChatRequest struct {
	BaseURL  string          `json:"baseUrl"`
	APIKey   string          `json:"apiKey,omitempty"`
	Model    string          `json:"model"`
	Messages []ChatMessage   `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ProxyChat forwards to an OpenAI-compatible endpoint (Ollama, LM Studio, etc.).
func ProxyChat(w http.ResponseWriter, r *http.Request, req ChatRequest) error {
	base, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return err
	}
	if req.Model == "" {
		return fmt.Errorf("model required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages required")
	}
	upstream := strings.TrimSuffix(base, "/") + "/chat/completions"
	body, _ := json.Marshal(map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	})
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstream, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(req.APIKey); key != "" {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	}
	client := &http.Client{Timeout: 0}
	if !req.Stream {
		client.Timeout = 120 * time.Second
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%w: %s", ErrUpstream, strings.TrimSpace(string(b)))
	}
	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, err := w.Write(buf[:n]); err != nil {
					return err
				}
				if ok {
					flusher.Flush()
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					return nil
				}
				return readErr
			}
		}
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidBaseURL
	}
	host := strings.ToLower(u.Hostname())
	allowed := map[string]bool{
		"127.0.0.1": true, "localhost": true, "::1": true,
		"ollama": true, "host.docker.internal": true,
	}
	if !allowed[host] {
		return "", ErrInvalidBaseURL
	}
	if !strings.HasSuffix(strings.TrimSuffix(raw, "/"), "/v1") {
		raw = strings.TrimSuffix(raw, "/") + "/v1"
	}
	return raw, nil
}
