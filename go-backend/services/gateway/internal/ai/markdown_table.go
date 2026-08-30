package ai

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var mdSepRE = regexp.MustCompile(`^\s*\|?[\s\-:|]+\|?\s*$`)

// parseMarkdownTable extracts header + data rows from a GitHub-flavored markdown table.
// Returns headers and rows (same width); skips empty lines outside the table.
func parseMarkdownTable(md string) (headers []string, rows [][]string, err error) {
	lines := strings.Split(md, "\n")
	var tableLines []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			if len(tableLines) > 0 {
				break
			}
			continue
		}
		if strings.Contains(trim, "|") {
			tableLines = append(tableLines, trim)
			continue
		}
		if len(tableLines) > 0 {
			break
		}
	}
	if len(tableLines) < 2 {
		return nil, nil, fmt.Errorf("no markdown table found (need header + separator + rows)")
	}
	headers = splitMDRow(tableLines[0])
	start := 1
	if mdSepRE.MatchString(tableLines[1]) {
		start = 2
	}
	if len(headers) == 0 {
		return nil, nil, fmt.Errorf("empty table header")
	}
	for _, line := range tableLines[start:] {
		if mdSepRE.MatchString(line) {
			continue
		}
		cells := splitMDRow(line)
		if len(cells) == 0 {
			continue
		}
		// Pad / trim to header width
		row := make([]string, len(headers))
		for i := range headers {
			if i < len(cells) {
				row[i] = cells[i]
			}
		}
		if rowAllEmpty(row) {
			continue
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("table has header but no data rows")
	}
	return headers, rows, nil
}

func splitMDRow(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")
	parts := strings.Split(s, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func rowAllEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func isSeqHeader(h string) bool {
	n := strings.TrimSpace(strings.ToLower(h))
	switch n {
	case "序号", "编号", "#", "no", "no.", "n", "index", "seq", "id":
		return true
	default:
		return false
	}
}

func guessFieldType(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(label, "日期") || strings.Contains(l, "date"):
		return "date"
	case strings.Contains(label, "数量") || strings.Contains(label, "单价") ||
		strings.Contains(label, "金额") || strings.Contains(label, "税额") ||
		strings.Contains(label, "税率") || strings.Contains(l, "amount") ||
		strings.Contains(l, "price") || strings.Contains(l, "qty"):
		return "number"
	default:
		return "text"
	}
}

func keyFromLabel(label string, used map[string]bool) string {
	base := knownLabelKey(label)
	if base == "" {
		base = slugifyLabel(label)
	}
	if base == "" {
		base = "col"
	}
	k := base
	for i := 2; used[k]; i++ {
		k = fmt.Sprintf("%s_%d", base, i)
	}
	used[k] = true
	return k
}

func knownLabelKey(label string) string {
	n := strings.TrimSpace(label)
	n = strings.ReplaceAll(n, " ", "")
	hints := map[string]string{
		"物品名称": "item_name", "名称": "item_name", "品名": "item_name", "商品": "item_name",
		"规格型号": "spec", "规格": "spec", "型号": "spec",
		"单位": "unit",
		"数量": "qty", "数量(个)": "qty",
		"含税单价(元)": "unit_price", "含税单价": "unit_price", "单价": "unit_price", "单价(元)": "unit_price",
		"金额(元)": "amount", "金额": "amount",
		"税率": "tax_rate",
		"税额(元)": "tax_amount", "税额": "tax_amount",
		"备注": "note", "说明": "note",
		"日期": "date", "时间": "date",
		"分类": "category", "用途": "note",
	}
	if k, ok := hints[n]; ok {
		return k
	}
	// strip parenthetical units e.g. 含税单价(元)
	if i := strings.IndexAny(n, "(（"); i > 0 {
		if k, ok := hints[n[:i]]; ok {
			return k
		}
	}
	return ""
}

func slugifyLabel(label string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(label) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(unicode.ToLower(r))
		case r == '_' || r == '-':
			b.WriteByte('_')
		case unicode.IsSpace(r):
			b.WriteByte('_')
		}
	}
	s := strings.Trim(b.String(), "_")
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	if s == "" {
		// Chinese-only: stable short key from runes
		runes := []rune(strings.TrimSpace(label))
		if len(runes) > 6 {
			runes = runes[:6]
		}
		return fmt.Sprintf("f_%x", []byte(string(runes)))
	}
	return s
}

// buildFieldPlan maps table headers → schema field keys; skips 序号.
type fieldPlan struct {
	HeaderIdx int
	Key       string
	Label     string
	Type      string
}

func planFieldsFromHeaders(headers []string) []fieldPlan {
	used := map[string]bool{}
	var plan []fieldPlan
	for i, h := range headers {
		if isSeqHeader(h) {
			continue
		}
		label := strings.TrimSpace(h)
		if label == "" {
			continue
		}
		plan = append(plan, fieldPlan{
			HeaderIdx: i,
			Key:       keyFromLabel(label, used),
			Label:     label,
			Type:      guessFieldType(label),
		})
	}
	return plan
}

func planFieldsFromSchema(headers []string, schemaFields []map[string]any) []fieldPlan {
	type sf struct {
		Key, Label string
	}
	var fields []sf
	for _, f := range schemaFields {
		fields = append(fields, sf{
			Key:   strings.TrimSpace(fmt.Sprint(f["key"])),
			Label: strings.TrimSpace(fmt.Sprint(f["label"])),
		})
	}
	var plan []fieldPlan
	usedKey := map[string]bool{}
	for i, h := range headers {
		if isSeqHeader(h) {
			continue
		}
		label := strings.TrimSpace(h)
		if label == "" {
			continue
		}
		key := ""
		norm := strings.ReplaceAll(label, " ", "")
		for _, f := range fields {
			if f.Key == "" || usedKey[f.Key] {
				continue
			}
			fl := strings.ReplaceAll(f.Label, " ", "")
			if fl == norm || f.Label == label || f.Key == label {
				key = f.Key
				break
			}
		}
		if key == "" {
			// fuzzy: label contains / contained
			for _, f := range fields {
				if f.Key == "" || usedKey[f.Key] {
					continue
				}
				fl := strings.ReplaceAll(f.Label, " ", "")
				if strings.Contains(fl, norm) || strings.Contains(norm, fl) {
					key = f.Key
					break
				}
			}
		}
		if key == "" {
			continue
		}
		usedKey[key] = true
		plan = append(plan, fieldPlan{HeaderIdx: i, Key: key, Label: label, Type: "text"})
	}
	return plan
}

func rowsToDataMaps(rows [][]string, plan []fieldPlan) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		m := map[string]any{}
		for _, p := range plan {
			val := ""
			if p.HeaderIdx < len(row) {
				val = strings.TrimSpace(row[p.HeaderIdx])
			}
			m[p.Key] = val
		}
		out = append(out, m)
	}
	return out
}
