package aiproxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
)

const agentMaxIterations = 6

// RunLedgerAgent executes an OpenAI Functions agent with ledger tools.
func RunLedgerAgent(ctx context.Context, llm llms.Model, backend LedgerToolBackend, userID, boundLedgerID, input, systemPrefix string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty agent input")
	}
	toolList := NewLedgerTools(backend, userID, boundLedgerID)
	opts := []agents.Option{agents.WithMaxIterations(agentMaxIterations)}
	if prefix := strings.TrimSpace(systemPrefix); prefix != "" {
		openAI := agents.NewOpenAIOption()
		opts = append(opts, openAI.WithSystemMessage(prefix))
	}
	agent := agents.NewOpenAIFunctionsAgent(llm, toolList, opts...)
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(agentMaxIterations))
	return chains.Run(ctx, executor, input)
}

func buildAgentInput(messages []ChatMessage) string {
	const maxTurns = 10
	var turns []string
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		content := strings.TrimSpace(m.Content)
		if role == "system" || content == "" {
			continue
		}
		label := "User"
		if role == "assistant" {
			label = "Assistant"
		}
		turns = append(turns, fmt.Sprintf("%s: %s", label, content))
	}
	if len(turns) == 0 {
		return ""
	}
	if len(turns) > maxTurns {
		turns = turns[len(turns)-maxTurns:]
	}
	last := turns[len(turns)-1]
	if !strings.HasPrefix(last, "User: ") {
		return last
	}
	if len(turns) == 1 {
		return strings.TrimPrefix(last, "User: ")
	}
	var b strings.Builder
	b.WriteString("Conversation:\n")
	for _, t := range turns[:len(turns)-1] {
		b.WriteString(t)
		b.WriteString("\n")
	}
	b.WriteString("\nCurrent question:\n")
	b.WriteString(strings.TrimPrefix(last, "User: "))
	return b.String()
}

func buildAgentSystemPrefix(messages []ChatMessage, boundLedgerID string) string {
	var parts []string
	for _, m := range messages {
		if strings.ToLower(strings.TrimSpace(m.Role)) != "system" {
			continue
		}
		c := strings.TrimSpace(m.Content)
		if c == "" {
			continue
		}
		if strings.Contains(c, "账本上下文") || strings.Contains(c, "RAG 导出") {
			parts = append(parts, "用户已绑定账本；请用 search_ledger_rag / get_ledger_summary 等工具查询最新链上数据。")
			continue
		}
		if len(c) > 4000 {
			c = c[:4000] + "…"
		}
		parts = append(parts, c)
	}
	if len(parts) == 0 {
		if p := DefaultAgentSystemPrompt(); p != "" {
			parts = append(parts, p)
		}
	}
	if id := strings.TrimSpace(boundLedgerID); id != "" {
		parts = append(parts, fmt.Sprintf("当前绑定账本 ID：%s。调用需要 ledgerId 的工具时，若用户未指定其他账本，请使用该 ID。", id))
	}
	parts = append(parts, "你可以通过 list_ledgers、get_ledger_summary、search_ledger_rag、verify_ledger 工具查询账本；不要编造链上数据。")
	return strings.Join(parts, "\n\n")
}
