package importfile

import (
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
)

// IsSeqHeader reports index/serial columns that should not become sheet fields.
func IsSeqHeader(h string) bool {
	n := strings.TrimSpace(strings.ToLower(h))
	n = strings.ReplaceAll(n, " ", "")
	switch n {
	case "序号", "编号", "#", "no", "no.", "n", "index", "seq", "id":
		return true
	default:
		return false
	}
}

// FilterSeqColumns drops 序号/# columns from headers and aligned row cells.
func FilterSeqColumns(headers []string, rows [][]string) ([]string, [][]string) {
	keep := make([]int, 0, len(headers))
	outH := make([]string, 0, len(headers))
	for i, h := range headers {
		if IsSeqHeader(h) {
			continue
		}
		if strings.TrimSpace(h) == "" {
			continue
		}
		keep = append(keep, i)
		outH = append(outH, strings.TrimSpace(h))
	}
	outRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		nr := make([]string, len(keep))
		for j, idx := range keep {
			if idx < len(row) {
				nr[j] = strings.TrimSpace(row[idx])
			}
		}
		empty := true
		for _, c := range nr {
			if c != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		outRows = append(outRows, nr)
	}
	return outH, outRows
}

// ParseCSVBytes reads CSV bytes into headers + data rows (header required).
func ParseCSVBytes(data []byte) ([]string, [][]string, error) {
	return readCSVRows(data)
}

// MapRowsByHeaders maps CSV header names onto an existing schema (label or key).
// Unmatched columns are ignored; missing schema fields get empty string.
func MapRowsByHeaders(headers []string, rows [][]string, schema domain.EntrySchema) []importxlsx.RowPreview {
	schema = domain.ResolveEntrySchema(schema)
	colToKey := map[int]string{}
	used := map[string]bool{}
	for i, h := range headers {
		label := strings.TrimSpace(h)
		if label == "" || IsSeqHeader(label) {
			continue
		}
		key := matchSchemaField(label, schema, used)
		if key == "" {
			continue
		}
		colToKey[i] = key
		used[key] = true
	}
	out := make([]importxlsx.RowPreview, 0, len(rows))
	line := 1
	for _, row := range rows {
		line++
		cells := map[string]string{}
		for _, f := range schema.Fields {
			cells[f.Key] = ""
		}
		for i, key := range colToKey {
			if i < len(row) {
				cells[key] = strings.TrimSpace(row[i])
			}
		}
		p := importxlsx.RowPreview{Line: line, Cells: cells}
		if err := domain.ValidateEntryData(schema, cells); err != nil {
			p.Error = err.Error()
		}
		out = append(out, p)
	}
	return out
}

func matchSchemaField(label string, schema domain.EntrySchema, used map[string]bool) string {
	norm := normLabel(label)
	for _, f := range schema.Fields {
		if used[f.Key] {
			continue
		}
		if normLabel(f.Label) == norm || normLabel(f.Key) == norm {
			return f.Key
		}
	}
	for _, f := range schema.Fields {
		if used[f.Key] {
			continue
		}
		fl := normLabel(f.Label)
		if strings.Contains(fl, norm) || strings.Contains(norm, fl) {
			return f.Key
		}
	}
	return ""
}

func normLabel(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	// strip unit parentheses: 金额(元) / 金额（元）
	if i := strings.IndexAny(s, "(（"); i > 0 {
		s = s[:i]
	}
	return s
}

// BuildImportSchemaFromHeaders builds a sheet schema; skips 序号; fields not required (append-friendly).
func BuildImportSchemaFromHeaders(headers []string) (domain.EntrySchema, error) {
	filtered, _ := FilterSeqColumns(headers, nil)
	if len(filtered) == 0 {
		return domain.EntrySchema{}, fmt.Errorf("no usable columns after skipping 序号")
	}
	schema := SchemaFromHeaders(filtered)
	// Import-friendly: do not force required (AI/CSV often has blank optional cells)
	for i := range schema.Fields {
		schema.Fields[i].Required = false
		// 税率 keep as text so "13%" validates
		if strings.Contains(schema.Fields[i].Label, "税率") {
			schema.Fields[i].Type = domain.FieldText
		}
	}
	schema.TemplateID = "import"
	return schema, nil
}

// PrepareCSVForNewSheet filters seq columns and builds schema + positional row previews.
func PrepareCSVForNewSheet(data []byte) (schema domain.EntrySchema, rows []importxlsx.RowPreview, err error) {
	headers, raw, err := readCSVRows(data)
	if err != nil {
		return schema, nil, err
	}
	headers, raw = FilterSeqColumns(headers, raw)
	schema, err = BuildImportSchemaFromHeaders(headers)
	if err != nil {
		return schema, nil, err
	}
	rows, err = importxlsx.ParseRowsWithSchema(schema, raw, 1)
	return schema, rows, err
}

// PrepareCSVForExistingSheet maps CSV columns onto an existing table schema by header name.
func PrepareCSVForExistingSheet(data []byte, schema domain.EntrySchema) ([]importxlsx.RowPreview, error) {
	headers, raw, err := readCSVRows(data)
	if err != nil {
		return nil, err
	}
	rows := MapRowsByHeaders(headers, raw, schema)
	if len(rows) == 0 {
		return nil, importxlsx.ErrNoDataRows
	}
	return rows, nil
}
