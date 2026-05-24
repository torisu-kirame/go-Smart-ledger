// Package miniledgerclient wraps Chainscore MiniLedger REST API.
// See https://github.com/Chainscore/miniledger
package miniledgerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client talks to a running MiniLedger node (default http://127.0.0.1:4441).
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:4441"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Status struct {
	Height uint64 `json:"height"`
	Uptime string `json:"uptime"`
	Role   string `json:"role,omitempty"`
}

type TxRequest struct {
	Key     string          `json:"key,omitempty"`
	Value   json.RawMessage `json:"value,omitempty"`
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type QueryRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params,omitempty"`
}

type StateRow struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("miniledger /status: %s (%s)", resp.Status, string(body))
	}
	return nil
}

func (c *Client) Status(ctx context.Context) (*Status, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("miniledger /status: %s (%s)", resp.Status, string(raw))
	}
	return parseStatus(raw)
}

func parseStatus(raw []byte) (*Status, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	st := &Status{}
	if v, ok := m["height"]; ok {
		st.Height = parseUintField(v)
	}
	if v, ok := m["uptime"]; ok {
		_ = json.Unmarshal(v, &st.Uptime)
	}
	if v, ok := m["role"]; ok {
		_ = json.Unmarshal(v, &st.Role)
	}
	return st, nil
}

func parseUintField(v json.RawMessage) uint64 {
	var n uint64
	if json.Unmarshal(v, &n) == nil {
		return n
	}
	var s string
	if json.Unmarshal(v, &s) == nil {
		u, _ := strconv.ParseUint(s, 10, 64)
		return u
	}
	var f float64
	if json.Unmarshal(v, &f) == nil {
		return uint64(f)
	}
	return 0
}

// Submit posts a transaction to MiniLedger.
func (c *Client) Submit(ctx context.Context, tx TxRequest) error {
	_, err := c.postJSON(ctx, "/tx", tx, nil)
	return err
}


// Query runs SQL against world_state.
func (c *Client) Query(ctx context.Context, sql string, params ...interface{}) ([]StateRow, error) {
	body := QueryRequest{SQL: sql, Params: params}
	var out struct {
		Rows    []StateRow `json:"rows"`
		Results []StateRow `json:"results"`
	}
	_, err := c.postJSON(ctx, "/state/query", body, &out)
	if err != nil {
		// some versions return array directly
		var rows []StateRow
		if _, err2 := c.postJSON(ctx, "/state/query", body, &rows); err2 == nil {
			return rows, nil
		}
		return nil, err
	}
	if len(out.Results) > 0 {
		return out.Results, nil
	}
	return out.Rows, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func (c *Client) postJSON(ctx context.Context, path string, body, out interface{}) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("miniledger %s: %s", path, string(raw))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return raw, err
		}
	}
	return raw, nil
}

func decodeResponse(resp *http.Response, out interface{}) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("miniledger: %s", string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// GetRaw performs GET and returns the response body (for proxying explorer APIs).
func (c *Client) GetRaw(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("miniledger %s: %s", path, string(raw))
	}
	return raw, nil
}

// BaseURL returns the configured node URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}
