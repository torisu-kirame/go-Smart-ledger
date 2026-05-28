package importxlsx

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/xuri/excelize/v2"
)

// TableSheetPreview is import preview for one ledger table / worksheet.
type TableSheetPreview struct {
	TableID     string       `json:"tableId"`
	TableName   string       `json:"tableName"`
	SheetName   string       `json:"sheetName,omitempty"`
	Rows        []RowPreview `json:"rows"`
	Valid       int          `json:"valid"`
	Invalid     int          `json:"invalid"`
	EntrySchema domain.EntrySchema `json:"entrySchema"`
}

// BuildTemplateWorkbook creates one sheet per table (sheet name = table name).
func BuildTemplateWorkbook(tables []domain.LedgerTable) ([]byte, error) {
	if len(tables) == 0 {
		return BuildTemplate(domain.DefaultEntrySchema())
	}
	f := excelize.NewFile()
	defer f.Close()
	first := true
	for _, t := range tables {
		sheet := sanitizeSheetName(t.Name)
		if first {
			_ = f.SetSheetName("Sheet1", sheet)
			first = false
		} else {
			_, _ = f.NewSheet(sheet)
		}
		schema := domain.ResolveEntrySchema(t.EntrySchema)
		for ci, fdef := range schema.Fields {
			cell, _ := excelize.CoordinatesToCellName(ci+1, 1)
			_ = f.SetCellValue(sheet, cell, fdef.Label)
		}
	}
	if len(tables) > 0 {
		if idx, err := f.GetSheetIndex(sanitizeSheetName(tables[0].Name)); err == nil {
			f.SetActiveSheet(idx)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseSheet reads one worksheet by name (empty = first sheet).
func ParseSheet(data []byte, sheetName string, schema domain.EntrySchema) ([]RowPreview, error) {
	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()
	sheet := sheetName
	if sheet == "" {
		sheet = f.GetSheetName(0)
	} else {
		found := false
		for _, n := range f.GetSheetList() {
			if n == sheetName {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("worksheet not found: %s", sheetName)
		}
	}
	if sheet == "" {
		return nil, ErrEmptyFile
	}
	return parseSheetRows(f, sheet, schema)
}

// ParseForTables maps worksheets to ledger tables by sheet name = table name.
func ParseForTables(data []byte, tables []domain.LedgerTable) ([]TableSheetPreview, error) {
	if len(tables) == 0 {
		return nil, fmt.Errorf("no tables")
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	byName := map[string]string{}
	for _, s := range sheets {
		byName[normSheetKey(s)] = s
	}
	var out []TableSheetPreview
	for _, t := range tables {
		sheet := byName[normSheetKey(t.Name)]
		if sheet == "" && len(tables) == 1 && len(sheets) > 0 {
			sheet = sheets[0]
		}
		schema := domain.ResolveEntrySchema(t.EntrySchema)
		p := TableSheetPreview{
			TableID:     t.ID,
			TableName:   t.Name,
			EntrySchema: schema,
		}
		if sheet == "" {
			out = append(out, p)
			continue
		}
		p.SheetName = sheet
		rows, err := parseSheetRows(f, sheet, schema)
		if err != nil {
			return nil, err
		}
		p.Rows = rows
		for _, row := range rows {
			if row.Error == "" {
				p.Valid++
			} else {
				p.Invalid++
			}
		}
		out = append(out, p)
	}
	return out, nil
}

func parseSheetRows(f *excelize.File, sheet string, schema domain.EntrySchema) ([]RowPreview, error) {
	schema = domain.ResolveEntrySchema(schema)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, ErrNoDataRows
	}
	header := map[string]int{}
	for i, h := range rows[0] {
		header[normHeader(h)] = i
	}
	labelToKey := map[string]string{}
	for _, fdef := range schema.Fields {
		labelToKey[normHeader(fdef.Label)] = fdef.Key
	}
	for _, fdef := range schema.Fields {
		if _, ok := header[normHeader(fdef.Label)]; !ok && fdef.Required {
			return nil, fmt.Errorf("missing column: %s", fdef.Label)
		}
	}
	var out []RowPreview
	for line := 1; line < len(rows); line++ {
		r := rows[line]
		if rowEmpty(r) {
			continue
		}
		if len(out) >= maxRows {
			return nil, ErrTooManyRows
		}
		cells := map[string]string{}
		for label, key := range labelToKey {
			if idx, ok := header[normHeader(label)]; ok {
				cells[key] = cell(r, idx)
			}
		}
		if schema.TemplateID == domain.TemplateClassic {
			if cells["type"] != "" {
				cells["type"] = normType(cells["type"])
			}
		}
		p := RowPreview{Line: line + 1, Cells: cells}
		fillLegacyFlat(&p)
		if err := validateRow(schema, &p); err != nil {
			p.Error = err.Error()
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, ErrNoDataRows
	}
	return out, nil
}

func sanitizeSheetName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Sheet"
	}
	repl := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "?", "-", "*", "-", "[", "(", "]", ")")
	name = repl.Replace(name)
	if len(name) > 31 {
		name = name[:31]
	}
	return name
}

func normSheetKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
