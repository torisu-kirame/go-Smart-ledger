package aiproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrOpenClawGateway = errors.New("openclaw gateway unreachable")
	ErrOpenClawTest    = errors.New("openclaw connection test failed")
)

// OpenClawRequest is sent from the console to proxy via OpenClaw Gateway.
type OpenClawRequest struct {
	GatewayURL     string          `json:"gatewayUrl"`
	GatewayToken   string          `json:"gatewayToken"`
	OpenClawModel  string          `json:"openclawModel,omitempty"`
	AgentUser      string          `json:"agentUser,omitempty"`
	Messages       []ChatMessage   `json:"messages"`
	Stream         bool            `json:"stream"`
	OpenClawConfig json.RawMessage `json:"openclawConfig,omitempty"`
}

// TestOpenClawRequest tests gateway + model routing.
type TestOpenClawRequest struct {
	GatewayURL     string          `json:"gatewayUrl"`
	GatewayToken   string          `json:"gatewayToken"`
	OpenClawConfig json.RawMessage `json:"openclawConfig,omitempty"`
}

// TestOpenClaw checks health and a minimal chat completion via OpenClaw.
func TestOpenClaw(r *http.Request, req TestOpenClawRequest) error {
	gw, err := effectiveGatewayURL(req.GatewayURL)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(req.GatewayToken)
	if token == "" {
		return fmt.Errorf("gateway token required")
	}
	if len(req.OpenClawConfig) > 0 {
		if err := writeOpenClawConfigIfPossible(req.OpenClawConfig); err != nil {
			return fmt.Errorf("failed to write openclaw config: %w", err)
		}
	}
	if err := pingGateway(r, gw); err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawGateway, err)
	}
	chatReq := OpenClawRequest{
		GatewayURL:    gw,
		GatewayToken:  token,
		OpenClawModel: "openclaw/default",
		Messages: []ChatMessage{
			{Role: "user", Content: "ping"},
		},
		Stream: false,
	}
	if err := proxyOpenClawOnce(r, chatReq); err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawTest, err)
	}
	return nil
}

// ProxyOpenClawChat forwards chat to OpenClaw Gateway /v1/chat/completions.
func ProxyOpenClawChat(w http.ResponseWriter, r *http.Request, req OpenClawRequest) error {
	gw, err := effectiveGatewayURL(req.GatewayURL)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(req.GatewayToken)
	if token == "" {
		return fmt.Errorf("gateway token required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages required")
	}
	model := strings.TrimSpace(req.OpenClawModel)
	if model == "" {
		model = "openclaw/default"
	}
	upstream := strings.TrimSuffix(gw, "/") + "/v1/chat/completions"
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": req.Messages,
		"stream":   req.Stream,
		"user":     strings.TrimSpace(req.AgentUser),
	})
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstream, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 0}
	if !req.Stream {
		client.Timeout = 180 * time.Second
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

func proxyOpenClawOnce(r *http.Request, req OpenClawRequest) error {
	gw, err := effectiveGatewayURL(req.GatewayURL)
	if err != nil {
		return err
	}
	model := strings.TrimSpace(req.OpenClawModel)
	if model == "" {
		model = "openclaw/default"
	}
	upstream := strings.TrimSuffix(gw, "/") + "/v1/chat/completions"
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": req.Messages,
		"stream":   false,
	})
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstream, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(req.GatewayToken))
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

func pingGateway(r *http.Request, gw string) error {
	healthURL := strings.TrimSuffix(gw, "/") + "/healthz"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("healthz status %d", resp.StatusCode)
	}
	return nil
}

func normalizeGatewayURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", ErrInvalidBaseURL
	}
	host := strings.ToLower(u.Hostname())
	allowed := map[string]bool{
		"127.0.0.1": true, "localhost": true, "::1": true,
		"openclaw-gateway": true, "host.docker.internal": true,
	}
	if !allowed[host] {
		return "", ErrInvalidBaseURL
	}
	return strings.TrimSuffix(raw, "/"), nil
}

// effectiveGatewayURL maps host localhost to OPENCLAW_GATEWAY_INTERNAL inside Docker.
func effectiveGatewayURL(raw string) (string, error) {
	gw, err := normalizeGatewayURL(raw)
	if err != nil {
		return "", err
	}
	internal := strings.TrimSpace(os.Getenv("OPENCLAW_GATEWAY_INTERNAL"))
	if internal == "" {
		return gw, nil
	}
	u, err := url.Parse(gw)
	if err != nil {
		return gw, nil
	}
	host := strings.ToLower(u.Hostname())
	if host == "127.0.0.1" || host == "localhost" || host == "::1" {
		if igw, ierr := normalizeGatewayURL(internal); ierr == nil {
			return igw, nil
		}
	}
	return gw, nil
}

func writeOpenClawConfigIfPossible(raw json.RawMessage) error {
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	ensureGatewayDefaults(cfg)
	ensureModelSchema(cfg)
	syncProviderAuth(cfg)
	merged, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	raw = merged
	path := strings.TrimSpace(os.Getenv("OPENCLAW_CONFIG_PATH"))
	if path == "" {
		path = filepath.Join("data", "openclaw", "config", "openclaw.json")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		pretty.Write(raw)
	} else {
		pretty.WriteByte('\n')
	}
	if err := os.WriteFile(path, pretty.Bytes(), 0o644); err != nil {
		return err
	}
	return writeAuthProfiles(dir, extractProviderKeys(cfg))
}

func ensureGatewayDefaults(cfg map[string]any) {
	gw, ok := cfg["gateway"].(map[string]any)
	if !ok || gw == nil {
		gw = map[string]any{}
		cfg["gateway"] = gw
	}
	if _, ok := gw["mode"]; !ok {
		gw["mode"] = "local"
	}
	if _, ok := gw["bind"]; !ok {
		gw["bind"] = "lan"
	}
	if _, ok := gw["port"]; !ok {
		gw["port"] = float64(18789)
	}
	httpBlock, ok := gw["http"].(map[string]any)
	if !ok || httpBlock == nil {
		httpBlock = map[string]any{}
		gw["http"] = httpBlock
	}
	endpoints, ok := httpBlock["endpoints"].(map[string]any)
	if !ok || endpoints == nil {
		endpoints = map[string]any{}
		httpBlock["endpoints"] = endpoints
	}
	chat, ok := endpoints["chatCompletions"].(map[string]any)
	if !ok || chat == nil {
		chat = map[string]any{"enabled": true}
		endpoints["chatCompletions"] = chat
	} else if _, ok := chat["enabled"]; !ok {
		chat["enabled"] = true
	}
	if auth, ok := gw["auth"].(map[string]any); ok && auth != nil {
		delete(auth, "token")
		if len(auth) == 0 {
			delete(gw, "auth")
		}
	}
	if _, ok := gw["controlUi"]; !ok {
		gw["controlUi"] = map[string]any{
			"allowedOrigins": []any{
				"http://localhost:18789",
				"http://127.0.0.1:18789",
				"http://localhost:25173",
				"http://127.0.0.1:25173",
			},
		}
	}
	ensureAgentTimeout(cfg)
}

func ensureAgentTimeout(cfg map[string]any) {
	agents, ok := cfg["agents"].(map[string]any)
	if !ok || agents == nil {
		return
	}
	defaults, ok := agents["defaults"].(map[string]any)
	if !ok || defaults == nil {
		return
	}
	if _, ok := defaults["timeoutSeconds"]; !ok {
		defaults["timeoutSeconds"] = float64(180)
	}
}

func ensureModelSchema(cfg map[string]any) {
	agents, ok := cfg["agents"].(map[string]any)
	if !ok || agents == nil {
		return
	}
	defaults, ok := agents["defaults"].(map[string]any)
	if !ok || defaults == nil {
		return
	}
	model, ok := defaults["model"].(map[string]any)
	if !ok || model == nil {
		return
	}
	if _, hasPrimary := model["primary"]; hasPrimary {
		return
	}
	provider, _ := model["provider"].(string)
	if provider == "" {
		return
	}
	baseURL, _ := model["baseUrl"].(string)
	if provider == "openai" && strings.Contains(baseURL, "deepseek") {
		provider = "deepseek"
	}
	modelID, _ := model["id"].(string)
	if modelID == "" {
		modelID = "deepseek-chat"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if provider == "deepseek" && strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	if baseURL == "" {
		if provider == "deepseek" {
			baseURL = "https://api.deepseek.com"
		} else {
			baseURL = "https://api.openai.com/v1"
		}
	}
	apiKey, _ := model["apiKey"].(string)
	modelRef := provider + "/" + modelID
	cfg["models"] = map[string]any{
		"mode": "merge",
		"providers": map[string]any{
			provider: map[string]any{
				"baseUrl": baseURL,
				"apiKey":  apiKey,
				"api":     "openai-completions",
				"models": []any{
					map[string]any{
						"id":            modelID,
						"name":          modelID,
						"reasoning":     false,
						"input":         []any{"text"},
						"cost":          map[string]any{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0},
						"contextWindow": 128000,
						"maxTokens":     8192,
					},
				},
			},
		},
	}
	defaults["model"] = map[string]any{"primary": modelRef}
	defaults["models"] = map[string]any{modelRef: map[string]any{}}
	delete(cfg, "plugins")
}

var providerAPIKeyEnv = map[string]string{
	"deepseek": "DEEPSEEK_API_KEY",
	"openai":   "OPENAI_API_KEY",
	"qwen":     "DASHSCOPE_API_KEY",
	"moonshot": "MOONSHOT_API_KEY",
	"groq":     "GROQ_API_KEY",
	"ollama":   "OLLAMA_API_KEY",
	"lmstudio": "LMSTUDIO_API_KEY",
}

func extractProviderKeys(cfg map[string]any) map[string]string {
	out := map[string]string{}
	models, ok := cfg["models"].(map[string]any)
	if !ok || models == nil {
		return out
	}
	providers, ok := models["providers"].(map[string]any)
	if !ok || providers == nil {
		return out
	}
	for name, block := range providers {
		prov, ok := block.(map[string]any)
		if !ok || prov == nil {
			continue
		}
		key, _ := prov["apiKey"].(string)
		key = strings.TrimSpace(key)
		if key != "" {
			out[name] = key
		}
	}
	return out
}

func syncProviderAuth(cfg map[string]any) {
	keys := extractProviderKeys(cfg)
	if len(keys) == 0 {
		return
	}
	env, ok := cfg["env"].(map[string]any)
	if !ok || env == nil {
		env = map[string]any{}
		cfg["env"] = env
	}
	for provider, key := range keys {
		if envName, ok := providerAPIKeyEnv[provider]; ok {
			env[envName] = key
		}
	}
	auth, ok := cfg["auth"].(map[string]any)
	if !ok || auth == nil {
		auth = map[string]any{}
		cfg["auth"] = auth
	}
	profiles, ok := auth["profiles"].(map[string]any)
	if !ok || profiles == nil {
		profiles = map[string]any{}
		auth["profiles"] = profiles
	}
	order, ok := auth["order"].(map[string]any)
	if !ok || order == nil {
		order = map[string]any{}
		auth["order"] = order
	}
	for provider := range keys {
		profileID := provider + ":default"
		profiles[profileID] = map[string]any{
			"provider": provider,
			"mode":     "api_key",
		}
		order[provider] = []any{profileID}
	}
}

func writeAuthProfiles(configDir string, keys map[string]string) error {
	if len(keys) == 0 {
		return nil
	}
	agentDir := filepath.Join(configDir, "agents", "main", "agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		return err
	}
	profiles := map[string]any{
		"version":  float64(1),
		"profiles": map[string]any{},
	}
	profMap := profiles["profiles"].(map[string]any)
	for provider, key := range keys {
		profMap[provider+":default"] = map[string]any{
			"type":     "api_key",
			"provider": provider,
			"key":      key,
		}
	}
	raw, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	authPath := filepath.Join(agentDir, "auth-profiles.json")
	return os.WriteFile(authPath, raw, 0o600)
}
