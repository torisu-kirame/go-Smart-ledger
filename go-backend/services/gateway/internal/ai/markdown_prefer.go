package ai

import "strings"

// preferRicherMarkdown picks the markdown with more data rows.
// Models often truncate tool-call arguments to the last 1–2 rows; the original
// user message usually still has the full table.
func preferRicherMarkdown(fromTool, fromSource string) string {
	tool := strings.TrimSpace(fromTool)
	src := extractMarkdownTableBlock(fromSource)
	if src == "" {
		src = strings.TrimSpace(fromSource)
	}
	if src == "" {
		return tool
	}
	if tool == "" {
		if _, _, err := parseMarkdownTable(src); err == nil {
			return src
		}
		return tool
	}
	_, rowsTool, errTool := parseMarkdownTable(tool)
	_, rowsSrc, errSrc := parseMarkdownTable(src)
	if errSrc != nil {
		return tool
	}
	if errTool != nil {
		return src
	}
	if len(rowsSrc) > len(rowsTool) {
		return src
	}
	return tool
}

// extractMarkdownTableBlock returns the first markdown table region in text, or "".
func extractMarkdownTableBlock(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	start, end := -1, -1
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.Contains(trim, "|") {
			if start < 0 {
				start = i
			}
			end = i
			continue
		}
		if start >= 0 && !strings.Contains(trim, "|") {
			break
		}
	}
	if start < 0 || end < start {
		return ""
	}
	return strings.Join(lines[start:end+1], "\n")
}
