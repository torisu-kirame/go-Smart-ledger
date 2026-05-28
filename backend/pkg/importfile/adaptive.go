package importfile

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
)

// AdaptiveResult is a file-driven import preview with inferred schema.
type AdaptiveResult struct {
	ProposedTableName string              `json:"proposedTableName"`
	EntrySchema       domain.EntrySchema  `json:"entrySchema"`
	Rows              []importxlsx.RowPreview `json:"rows"`
	Valid             int                 `json:"valid"`
	Invalid           int                 `json:"invalid"`
	Total             int                 `json:"total"`
	FileKind          string              `json:"fileKind"` // xlsx | csv
}

// ParseAdaptive reads xlsx/csv and builds schema from header row.
func ParseAdaptive(data []byte, filename string) (*AdaptiveResult, error) {
	if len(data) == 0 {
		return nil, importxlsx.ErrEmptyFile
	}
	name := strings.TrimSpace(filename)
	var headers []string
	var rows [][]string
	var kind string
	var err error
	if isCSVName(name) {
		kind = "csv"
		headers, rows, err = readCSVRows(data)
	} else {
		kind = "xlsx"
		headers, rows, err = readXLSXFirstSheet(data)
	}
	if err != nil {
		return nil, err
	}
	schema := SchemaFromHeaders(headers)
	tableName := sanitizeTableName(TableNameFromFilename(name))
	parsed, err := importxlsx.ParseRowsWithSchema(schema, rows, 1)
	if err != nil {
		return nil, err
	}
	valid, invalid := 0, 0
	for _, r := range parsed {
		if r.Error == "" {
			valid++
		} else {
			invalid++
		}
	}
	return &AdaptiveResult{
		ProposedTableName: tableName,
		EntrySchema:       schema,
		Rows:              parsed,
		Valid:             valid,
		Invalid:           invalid,
		Total:             len(parsed),
		FileKind:          kind,
	}, nil
}

// SchemaFromHeaders builds entry schema from column titles.
func SchemaFromHeaders(headers []string) domain.EntrySchema {
	fields := make([]domain.EntryFieldDef, 0, len(headers))
	seen := map[string]bool{}
	for i, h := range headers {
		label := strings.TrimSpace(h)
		if label == "" {
			continue
		}
		key := inferFieldKey(label, i, seen)
		ft := inferFieldType(label)
		required := ft == domain.FieldDate || ft == domain.FieldNumber
		if strings.Contains(strings.ToLower(label), "记账人") || strings.Contains(strings.ToLower(label), "bookkeeper") {
			required = true
		}
		fields = append(fields, domain.EntryFieldDef{
			Key:      key,
			Label:    label,
			Type:     ft,
			Required: required,
		})
	}
	if len(fields) == 0 {
		fields = domain.DefaultEntrySchema().Fields
	}
	return domain.EntrySchema{
		TemplateID: "import",
		Fields:     fields,
	}
}

func inferFieldKey(label string, index int, seen map[string]bool) string {
	lower := strings.ToLower(strings.TrimSpace(label))
	exact := map[string]string{
		"date": "date", "amount": "amount", "type": "type",
		"category": "category", "note": "note", "bookkeeper": "bookkeeper",
		"payee": "payee", "counterparty": "counterparty",
	}
	if k, ok := exact[lower]; ok && !seen[k] {
		seen[k] = true
		return k
	}
	contains := []struct{ sub, key string }{
		{"日期", "date"}, {"记账日期", "date"},
		{"金额", "amount"}, {"数额", "amount"},
		{"类型", "type"}, {"收支", "type"},
		{"分类", "category"}, {"备注", "note"}, {"说明", "note"},
		{"记账人", "bookkeeper"}, {"收账人", "payee"}, {"对方", "counterparty"},
	}
	for _, m := range contains {
		if strings.Contains(label, m.sub) && !seen[m.key] {
			seen[m.key] = true
			return m.key
		}
	}
	base := "col_" + fmt.Sprintf("%d", index+1)
	key := base
	n := 1
	for seen[key] {
		key = fmt.Sprintf("%s_%d", base, n)
		n++
	}
	seen[key] = true
	return key
}

func inferFieldType(label string) domain.FieldType {
	lower := strings.ToLower(strings.TrimSpace(label))
	switch {
	case strings.Contains(lower, "日期") || lower == "date" || strings.Contains(lower, "时间"):
		return domain.FieldDate
	case strings.Contains(lower, "金额") || strings.Contains(lower, "amount") || strings.Contains(lower, "价格") || strings.Contains(lower, "数额"):
		return domain.FieldNumber
	case strings.Contains(lower, "记账人") || strings.Contains(lower, "bookkeeper") || strings.Contains(lower, "用户"):
		return domain.FieldUser
	default:
		return domain.FieldText
	}
}

// TableNameFromFilename derives a display table name from upload filename.
func TableNameFromFilename(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	base = strings.TrimSpace(base)
	if base == "" {
		return "导入数据"
	}
	runes := []rune(base)
	if len(runes) > 24 {
		base = string(runes[:24])
	}
	return base
}

// IsCSVName reports whether filename looks like CSV.
func IsCSVName(name string) bool {
	return isCSVName(name)
}

func isCSVName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".csv" || ext == ".txt"
}

func sanitizeTableName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "导入数据"
	}
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == ' ' {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "导入数据"
	}
	return out
}
