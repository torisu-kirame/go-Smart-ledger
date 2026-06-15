package aiproxy

import (
	"embed"
	"strings"
)

//go:embed workspace/*
var embeddedWorkspace embed.FS

// DefaultAgentSystemPrompt loads bundled AGENTS.md + TOOLS.md for LangChain system prompt.
func DefaultAgentSystemPrompt() string {
	var b strings.Builder
	for _, name := range []string{"workspace/AGENTS.md", "workspace/TOOLS.md"} {
		raw, err := embeddedWorkspace.ReadFile(name)
		if err != nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n\n---\n\n")
		}
		b.Write(raw)
	}
	return strings.TrimSpace(b.String())
}
