package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedWorkspaceFiles = map[string]bool{
	"AGENTS.md": true, "BOOTSTRAP.md": true, "HEARTBEAT.md": true,
	"IDENTITY.md": true, "SOUL.md": true, "TOOLS.md": true, "USER.md": true,
	"API-REFERENCE.md": true,
}

const chatMessagesFile = "messages.json"

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Storage struct {
	ConfigRoot string
}

func (s *Storage) resolveSafe(rel string) (string, error) {
	rel = strings.TrimSpace(strings.ReplaceAll(rel, `\`, "/"))
	if rel == "" || strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") {
		return "", fmt.Errorf("agent path not allowed: %s", rel)
	}
	base := filepath.Clean(s.ConfigRoot)
	full := filepath.Clean(filepath.Join(base, filepath.FromSlash(rel)))
	relOut, err := filepath.Rel(base, full)
	if err != nil || strings.HasPrefix(relOut, "..") {
		return "", fmt.Errorf("agent path not allowed: %s", rel)
	}
	return full, nil
}

func (s *Storage) workspaceDir(agentPath string) string {
	normalized := strings.Trim(agentPath, `/\`)
	if normalized == "agents/main" {
		legacy := filepath.Join(s.ConfigRoot, "workspace", "workspace-smart-ledger")
		if st, err := os.Stat(legacy); err == nil && st.IsDir() {
			return legacy
		}
	}
	return filepath.Join(s.ConfigRoot, filepath.FromSlash(normalized), "workspace")
}

func (s *Storage) Load(agentPath, chatHistoryPath string, loadWorkspace bool) (messages []ChatMessage, workspace map[string]string, updatedAt string, err error) {
	workspace = map[string]string{}
	chatDir, err := s.resolveSafe(chatHistoryPath)
	if err != nil {
		return nil, nil, "", err
	}
	chatPath := filepath.Join(chatDir, chatMessagesFile)
	if b, readErr := os.ReadFile(chatPath); readErr == nil {
		var payload struct {
			Messages  []ChatMessage `json:"messages"`
			UpdatedAt string        `json:"updatedAt"`
		}
		if json.Unmarshal(b, &payload) == nil {
			messages = payload.Messages
			updatedAt = payload.UpdatedAt
		}
	}
	if loadWorkspace {
		ws := s.workspaceDir(agentPath)
		for name := range allowedWorkspaceFiles {
			fp := filepath.Join(ws, name)
			if b, e := os.ReadFile(fp); e == nil {
				workspace[name] = string(b)
			}
		}
	}
	return messages, workspace, updatedAt, nil
}

func (s *Storage) Save(agentPath, chatHistoryPath string, messages []ChatMessage, workspace map[string]string) error {
	if len(messages) > 0 {
		chatDir, err := s.resolveSafe(chatHistoryPath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(chatDir, 0o755); err != nil {
			return err
		}
		payload := map[string]any{
			"version":   1,
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
			"messages":  messages,
		}
		raw, _ := json.MarshalIndent(payload, "", "  ")
		if err := os.WriteFile(filepath.Join(chatDir, chatMessagesFile), append(raw, '\n'), 0o644); err != nil {
			return err
		}
	}
	if len(workspace) > 0 {
		ws := s.workspaceDir(agentPath)
		if err := os.MkdirAll(ws, 0o755); err != nil {
			return err
		}
		for name, content := range workspace {
			if !allowedWorkspaceFiles[name] {
				continue
			}
			if err := os.WriteFile(filepath.Join(ws, name), []byte(content), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
