package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// autoImportMarkdownTable bypasses the LLM for multi-row Markdown table imports.
// Models routinely truncate tool-call arguments to 1 row; the gateway already has
// the full user text, so we import it server-side when intent is clear.
func autoImportMarkdownTable(ctx context.Context, ledger *LedgerHTTP, boundLedgerID, source string) (string, bool) {
	if ledger == nil || strings.TrimSpace(boundLedgerID) == "" {
		return "", false
	}
	n := sourceTableRowCount(source)
	if n < 2 {
		return "", false
	}
	if !looksLikeTableImportIntent(source) {
		return "", false
	}
	sheetName := guessImportSheetName(source)
	raw := ledger.toolImportMarkdown(ctx, map[string]any{
		"sheetName": sheetName,
		// markdown omitted on purpose — preferRicherMarkdown uses SourceText
	}, boundLedgerID)
	if strings.HasPrefix(strings.TrimSpace(raw), "error:") {
		return "", false
	}
	expected, imported, mode, tableID, tableName := summarizeImportToolResult(raw)
	var b strings.Builder
	b.WriteString("## 表格导入完成\n\n")
	b.WriteString(fmt.Sprintf("已从你消息中的 Markdown 表格解析 **%d** 行，并经批量导入 API 写入账本", expected))
	if expected > 0 && imported > 0 && imported != expected {
		b.WriteString(fmt.Sprintf("（实际成功 **%d** 行，请核对失败行）", imported))
	} else if imported > 0 {
		b.WriteString(fmt.Sprintf("（成功 **%d** 行）", imported))
	}
	b.WriteString("。\n\n")
	if mode != "" || tableName != "" || tableID != "" {
		b.WriteString("| 项 | 值 |\n| --- | --- |\n")
		if mode != "" {
			b.WriteString("| 模式 | `" + mode + "` |\n")
		}
		if tableName != "" {
			b.WriteString("| Sheet | " + tableName + " |\n")
		}
		if tableID != "" {
			b.WriteString("| tableId | `" + tableID + "` |\n")
		}
		b.WriteString(fmt.Sprintf("| 解析行数 | %d |\n", expected))
		b.WriteString(fmt.Sprintf("| 写入行数 | %d |\n", imported))
		b.WriteString("\n")
	}
	b.WriteString("本次由网关直接批量导入，**未**逐条调用 `append_ledger_entry`，避免模型截断表格。\n")
	return b.String(), true
}

func looksLikeTableImportIntent(s string) bool {
	if sourceTableRowCount(s) < 2 {
		return false
	}
	keys := []string{
		"导入", "添加", "录入", "写入", "加到", "放到", "插入", "同步到",
		"import", "append", "insert",
	}
	for _, k := range keys {
		if strings.Contains(s, k) {
			return true
		}
	}
	// Bare multi-row table (no chatter) with bound ledger → treat as import
	return sourceTableRowCount(s) >= 3
}

func guessImportSheetName(s string) string {
	lower := s
	switch {
	case strings.Contains(lower, "采购"):
		return "采购明细"
	case strings.Contains(lower, "库存"):
		return "库存明细"
	case strings.Contains(lower, "报销"):
		return "报销明细"
	default:
		return "导入明细"
	}
}

func summarizeImportToolResult(raw string) (expected, imported int, mode, tableID, tableName string) {
	var root map[string]any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return 0, 0, "", "", ""
	}
	if v, ok := root["expectedRows"].(float64); ok {
		expected = int(v)
	}
	imp := root["importResult"]
	var walk func(map[string]any)
	walk = func(m map[string]any) {
		if m == nil {
			return
		}
		if v, ok := m["mode"].(string); ok && mode == "" {
			mode = v
		}
		if v, ok := m["tableId"].(string); ok && tableID == "" {
			tableID = v
		}
		if v, ok := m["tableName"].(string); ok && tableName == "" {
			tableName = v
		}
		if v, ok := m["parsedRows"].(float64); ok && expected == 0 {
			expected = int(v)
		}
		if nest, ok := m["import"].(map[string]any); ok {
			if v, ok := nest["imported"].(float64); ok {
				imported = int(v)
			}
			if v, ok := nest["mode"].(string); ok && mode == "" {
				mode = v
			}
			if v, ok := nest["tableId"].(string); ok && tableID == "" {
				tableID = v
			}
			if v, ok := nest["tableName"].(string); ok && tableName == "" {
				tableName = v
			}
			if inner, ok := nest["import"].(map[string]any); ok {
				if v, ok := inner["imported"].(float64); ok {
					imported = int(v)
				}
			}
			walk(nest)
		}
		if res, ok := m["result"].(map[string]any); ok {
			walk(res)
		}
	}
	if m, ok := imp.(map[string]any); ok {
		walk(m)
	}
	walk(root)
	return expected, imported, mode, tableID, tableName
}
