package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const maxAgentIterations = 24

func runToolAgent(
	ctx context.Context,
	oc *OpenClawClient,
	ledger *LedgerHTTP,
	messages []map[string]any,
	boundLedgerID, openclawModel, agentModel, sessionKey string,
) (string, error) {
	// Multi-row Markdown tables: import server-side first (LLM tool args are unreliable).
	sourceText := ""
	if ledger != nil {
		sourceText = ledger.SourceText
	}
	if summary, ok := autoImportMarkdownTable(ctx, ledger, boundLedgerID, sourceText); ok {
		return summary, nil
	}

	model := strings.TrimSpace(agentModel)
	if model == "" {
		model = "openclaw/default"
	}
	msgs := append([]map[string]any{}, messages...)
	for i := 0; i < maxAgentIterations; i++ {
		body := map[string]any{
			"model":       model,
			"messages":    msgs,
			"tools":       toolDefinitions(),
			"tool_choice": "auto",
			"stream":      false,
		}
		// Fresh OpenClaw session every model call — sticky sessions + tool_calls
		// caused "Cannot continue from message role: assistant" / internal error.
		turnKey := strings.TrimSpace(sessionKey)
		if turnKey == "" {
			turnKey = "sl"
		}
		turnKey = fmt.Sprintf("%s-t%d-%s", turnKey, i, uuid.NewString()[:8])
		resp, err := oc.ChatCompletion(ctx, body, openclawModel, turnKey)
		if err != nil {
			return "", err
		}
		toolCalls := messageToolCalls(resp)
		content := messageContent(resp)
		choices, _ := resp["choices"].([]any)
		var msg map[string]any
		if len(choices) > 0 {
			ch, _ := choices[0].(map[string]any)
			msg, _ = ch["message"].(map[string]any)
		}
		if len(toolCalls) > 0 {
			assistant := map[string]any{
				"role":       "assistant",
				"content":    "",
				"tool_calls": toolCalls,
			}
			if msg != nil {
				if c, ok := msg["content"]; ok {
					assistant["content"] = c
				}
			}
			msgs = append(msgs, assistant)
			// MUST run tools sequentially. Parallel append_ledger_entry races on
			// the same ledger seq and silently overwrites events.
			for _, tc := range toolCalls {
				name, args, callID := parseToolCall(tc)
				result := ledger.invokeTool(ctx, name, args, boundLedgerID)
				msgs = append(msgs, map[string]any{
					"role":         "tool",
					"tool_call_id": callID,
					"content":      result,
				})
			}
			continue
		}
		if content != "" {
			return content, nil
		}
		break
	}
	return "Agent 已达到最大工具调用轮次仍未给出最终答复。可改用 call_ledger_api 或直接在控制台操作。", nil
}

func parseToolCall(tc map[string]any) (name string, args map[string]any, callID string) {
	args = map[string]any{}
	callID, _ = tc["id"].(string)
	fn, _ := tc["function"].(map[string]any)
	if fn == nil {
		return "", args, callID
	}
	name, _ = fn["name"].(string)
	switch a := fn["arguments"].(type) {
	case string:
		_ = json.Unmarshal([]byte(a), &args)
	case map[string]any:
		args = a
	}
	if args == nil {
		args = map[string]any{}
	}
	return name, args, callID
}

func buildAgentMessages(systemPrefix, userInput string) []map[string]any {
	sys := strings.TrimSpace(systemPrefix)
	if sys == "" {
		sys = "你是 Smart Ledger 智能账本助手。可通过 call_ledger_api 调用账本 REST。回复使用 Markdown。"
	}
	return []map[string]any{
		{"role": "system", "content": sys},
		{"role": "user", "content": userInput},
	}
}

func buildUserInput(messages []ChatMessage) string {
	var turns []string
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		c := strings.TrimSpace(m.Content)
		if role == "system" || c == "" {
			continue
		}
		label := "User"
		if role == "assistant" {
			label = "Assistant"
		}
		turns = append(turns, label+": "+c)
	}
	if len(turns) == 0 {
		return ""
	}
	if len(turns) > 10 {
		turns = turns[len(turns)-10:]
	}
	last := turns[len(turns)-1]
	if !strings.HasPrefix(last, "User: ") {
		return last
	}
	if len(turns) == 1 {
		return strings.TrimPrefix(last, "User: ")
	}
	return "Conversation:\n" + strings.Join(turns[:len(turns)-1], "\n") +
		"\n\nCurrent question:\n" + strings.TrimPrefix(last, "User: ")
}

// latestUserContent returns only the last user message (not conversation history).
// Auto-import / markdown prefer must use this — otherwise an earlier pasted table
// would re-trigger import on every follow-up ("你好", "为什么空白"…).
func latestUserContent(messages []ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		role := strings.ToLower(strings.TrimSpace(messages[i].Role))
		c := strings.TrimSpace(messages[i].Content)
		if role == "user" && c != "" {
			return c
		}
	}
	return ""
}

func buildSystemPrefix(messages []ChatMessage, boundLedgerID string) string {
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
			parts = append(parts, "用户已绑定账本；请用 search_ledger_rag / get_ledger_summary / call_ledger_api 查询最新数据。")
			continue
		}
		if len(c) > 4000 {
			c = c[:4000] + "…"
		}
		parts = append(parts, c)
	}
	if lid := strings.TrimSpace(boundLedgerID); lid != "" {
		parts = append(parts, fmt.Sprintf("当前绑定账本 ID：%s。调用工具时默认使用该 ID。", lid))
	}
	parts = append(parts,
		"你运行在 Smart Ledger → OpenClaw Agent 链路（gateway 编排工具）。",
		"优先专用工具：import_markdown_table / import_sheet_csv（MD→CSV→sheet-csv 批量导入）、create_ledger、create_ledger_sheet、"+
			"append_ledger_entries_batch、append_ledger_entry、list_ledgers、get_ledger_summary、search_ledger_rag；必要时再用 call_ledger_api。",
		"用户给出 Markdown 表格时：调用 import_markdown_table（markdown 可省略，服务端会用用户消息里的完整表）。"+
			"禁止对每一行单独调用 append_ledger_entry。tableId 空=新建 Sheet；有 tableId=追加到该 Sheet 底部。",
		"写分录 API 必须 body={\"entry\":{\"tableId\":\"...\",\"data\":{字段key:值}}}，不能把 tableId 放在顶层。",
		"建 Sheet：create_ledger_sheet（fields 含 key/label/type/required；不要建「序号」列）。",
		"不要编造链上数据。回复使用 Markdown。写完后核对 written 与 parsedRows 是否一致。",
	)
	return strings.Join(parts, "\n\n")
}

func resolveOpenClawModel(providerModel string) string {
	model := strings.TrimSpace(providerModel)
	if model == "" {
		return ""
	}
	if strings.Contains(model, "/") {
		return model
	}
	switch {
	case strings.HasPrefix(model, "deepseek"):
		return "deepseek/" + model
	case strings.HasPrefix(model, "gpt-"), strings.HasPrefix(model, "o1"), strings.HasPrefix(model, "o3"):
		return "openai/" + model
	case strings.HasPrefix(model, "qwen"):
		return "qwen/" + model
	default:
		return model
	}
}
