package aiproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrAgentPathInvalid = errors.New("agent path not allowed")
	ErrAgentStorage     = errors.New("agent storage error")
)

const chatMessagesFile = "messages.json"

var allowedWorkspaceFiles = map[string]bool{
	"AGENTS.md":    true,
	"BOOTSTRAP.md": true,
	"HEARTBEAT.md": true,
	"IDENTITY.md":  true,
	"SOUL.md":      true,
	"TOOLS.md":     true,
	"USER.md":      true,
}

// AgentStoragePaths locates agent data on disk (relative to agent config root).
type AgentStoragePaths struct {
	AgentPath       string `json:"agentPath"`
	ChatHistoryPath string `json:"chatHistoryPath"`
}

// AgentChatMessage is one turn in a conversation.
type AgentChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentStorageLoadRequest loads chat and optional workspace files.
type AgentStorageLoadRequest struct {
	AgentStoragePaths
	LoadWorkspace bool `json:"loadWorkspace"`
}

// AgentStorageLoadResponse is returned from disk load.
type AgentStorageLoadResponse struct {
	Messages       []AgentChatMessage  `json:"messages"`
	WorkspaceFiles map[string]string   `json:"workspaceFiles,omitempty"`
	UpdatedAt      string              `json:"updatedAt,omitempty"`
}

// AgentStorageSaveRequest persists chat and/or workspace files.
type AgentStorageSaveRequest struct {
	AgentStoragePaths
	Messages       []AgentChatMessage `json:"messages,omitempty"`
	WorkspaceFiles map[string]string  `json:"workspaceFiles,omitempty"`
}

func agentConfigRoot() string {
	path := strings.TrimSpace(os.Getenv("AGENT_CONFIG_PATH"))
	if path != "" {
		return path
	}
	path = strings.TrimSpace(os.Getenv("OPENCLAW_CONFIG_PATH"))
	if path != "" {
		return filepath.Dir(path)
	}
	return filepath.Join("data", "agent", "config")
}

func openClawConfigRoot() string {
	return agentConfigRoot()
}

func resolveSafeRelPath(baseDir, relPath string) (string, error) {
	relPath = strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if relPath == "" {
		return "", fmt.Errorf("%w: empty path", ErrAgentPathInvalid)
	}
	if filepath.IsAbs(relPath) || strings.Contains(relPath, "..") {
		return "", fmt.Errorf("%w: %s", ErrAgentPathInvalid, relPath)
	}
	clean := filepath.Clean(relPath)
	full := filepath.Join(baseDir, clean)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if absFull != absBase && !strings.HasPrefix(absFull, absBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: %s", ErrAgentPathInvalid, relPath)
	}
	return absFull, nil
}

func workspaceDirForAgent(configRoot, agentPath string) string {
	normalized := strings.Trim(strings.ReplaceAll(agentPath, "\\", "/"), "/")
	if normalized == "agents/main" {
		legacy := filepath.Join(configRoot, "workspace", "workspace-smart-ledger")
		if st, err := os.Stat(legacy); err == nil && st.IsDir() {
			return legacy
		}
		// Bundled defaults are embedded; editable copies live under agents/*/workspace.
	}
	return filepath.Join(configRoot, normalized, "workspace")
}

// LoadAgentStorage reads chat history and optional workspace markdown files.
func LoadAgentStorage(req AgentStorageLoadRequest) (*AgentStorageLoadResponse, error) {
	root := openClawConfigRoot()
	chatDir, err := resolveSafeRelPath(root, req.ChatHistoryPath)
	if err != nil {
		return nil, err
	}
	out := &AgentStorageLoadResponse{
		Messages:       []AgentChatMessage{},
		WorkspaceFiles: map[string]string{},
	}
	chatPath := filepath.Join(chatDir, chatMessagesFile)
	raw, err := os.ReadFile(chatPath)
	if err == nil {
		var payload struct {
			Messages  []AgentChatMessage `json:"messages"`
			UpdatedAt string             `json:"updatedAt"`
		}
		if json.Unmarshal(raw, &payload) == nil {
			if payload.Messages != nil {
				out.Messages = payload.Messages
			}
			out.UpdatedAt = payload.UpdatedAt
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: read chat: %v", ErrAgentStorage, err)
	}
	if req.LoadWorkspace {
		wsDir := workspaceDirForAgent(root, req.AgentPath)
		for name := range allowedWorkspaceFiles {
			b, readErr := os.ReadFile(filepath.Join(wsDir, name))
			if readErr == nil {
				out.WorkspaceFiles[name] = string(b)
			}
		}
	}
	return out, nil
}

// SaveAgentStorage writes chat history and/or workspace markdown files.
func SaveAgentStorage(req AgentStorageSaveRequest) error {
	root := openClawConfigRoot()
	if len(req.Messages) > 0 {
		chatDir, err := resolveSafeRelPath(root, req.ChatHistoryPath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(chatDir, 0o755); err != nil {
			return fmt.Errorf("%w: mkdir chat: %v", ErrAgentStorage, err)
		}
		payload, err := json.MarshalIndent(map[string]any{
			"version":   1,
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
			"messages":  req.Messages,
		}, "", "  ")
		if err != nil {
			return err
		}
		payload = append(payload, '\n')
		chatPath := filepath.Join(chatDir, chatMessagesFile)
		if err := os.WriteFile(chatPath, payload, 0o644); err != nil {
			return fmt.Errorf("%w: write chat: %v", ErrAgentStorage, err)
		}
	}
	if len(req.WorkspaceFiles) > 0 {
		wsDir := workspaceDirForAgent(root, req.AgentPath)
		if err := os.MkdirAll(wsDir, 0o755); err != nil {
			return fmt.Errorf("%w: mkdir workspace: %v", ErrAgentStorage, err)
		}
		for name, content := range req.WorkspaceFiles {
			if !allowedWorkspaceFiles[name] {
				continue
			}
			if err := os.WriteFile(filepath.Join(wsDir, name), []byte(content), 0o644); err != nil {
				return fmt.Errorf("%w: write %s: %v", ErrAgentStorage, name, err)
			}
		}
	}
	return nil
}
