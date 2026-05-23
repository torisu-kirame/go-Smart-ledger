// Package ipfsclient talks to Kubo (go-ipfs) HTTP API.
package ipfsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client is an IPFS Kubo HTTP API client.
type Client struct {
	apiURL     string
	httpClient *http.Client
}

func New(apiURL string) *Client {
	apiURL = strings.TrimRight(apiURL, "/")
	return &Client{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiURL != ""
}

// Add uploads bytes and returns CID (with pin when pin=true).
func (c *Client) Add(ctx context.Context, data []byte, pin bool) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("ipfs client disabled")
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "backup.bin")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	pinQ := "false"
	if pin {
		pinQ = "true"
	}
	url := fmt.Sprintf("%s/api/v0/add?pin=%s&quieter=true", c.apiURL, pinQ)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ipfs add: %s", string(b))
	}
	dec := json.NewDecoder(resp.Body)
	var last struct {
		Hash string `json:"Hash"`
		Name string `json:"Name"`
	}
	for dec.More() {
		if err := dec.Decode(&last); err != nil {
			return "", err
		}
	}
	if last.Hash == "" {
		return "", fmt.Errorf("ipfs add: empty cid")
	}
	return last.Hash, nil
}

// Pin ensures a CID stays pinned.
func (c *Client) Pin(ctx context.Context, cid string) error {
	if !c.Enabled() {
		return fmt.Errorf("ipfs client disabled")
	}
	url := fmt.Sprintf("%s/api/v0/pin/add?arg=%s", c.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ipfs pin: %s", string(b))
	}
	return nil
}

// Cat reads content by CID.
func (c *Client) Cat(ctx context.Context, cid string) ([]byte, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("ipfs client disabled")
	}
	url := fmt.Sprintf("%s/api/v0/cat?arg=%s", c.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ipfs cat: %s", string(b))
	}
	return io.ReadAll(resp.Body)
}

// Ping checks API reachability.
func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("ipfs disabled")
	}
	url := fmt.Sprintf("%s/api/v0/id", c.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ipfs id: status %d", resp.StatusCode)
	}
	return nil
}
