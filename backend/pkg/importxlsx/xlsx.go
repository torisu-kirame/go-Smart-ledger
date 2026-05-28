package importxlsx

import (
	"bytes"
	"errors"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/xuri/excelize/v2"
)

var (
	ErrEmptyFile   = errors.New("empty excel file")
	ErrNoDataRows  = errors.New("no data rows found")
	ErrTooManyRows = errors.New("too many rows (max 5000)")
)

const maxRows = 5000

// RowPreview is one parsed row with dynamic columns in Cells.
type RowPreview struct {
	Line   int               `json:"line"`
	Cells  map[string]string `json:"cells"`
	Error  string            `json:"error,omitempty"`
	// Legacy flat fields for older frontends
	Date         string `json:"date,omitempty"`
	Type         string `json:"type,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Category     string `json:"category,omitempty"`
	Note         string `json:"note,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
}

// Parse reads xlsx using ledger entry schema column labels.
func Parse(data []byte, schema domain.EntrySchema) ([]RowPreview, error) {
	schema = domain.ResolveEntrySchema(schema)
	return parseWithSchema(data, schema)
}

// ParseLegacy uses classic 日期/类型/金额 columns.
func ParseLegacy(data []byte) ([]RowPreview, error) {
	return parseWithSchema(data, domain.ClassicEntrySchema())
}

func parseWithSchema(data []byte, schema domain.EntrySchema) ([]RowPreview, error) {
	return ParseSheet(data, "", schema)
}

// ToEntry converts valid preview row to domain entry.
func ToEntry(p RowPreview, schema domain.EntrySchema) (domain.EntryPayload, error) {
	if p.Error != "" {
		return domain.EntryPayload{}, errors.New(p.Error)
	}
	data := p.Cells
	if len(data) == 0 {
		data = map[string]string{
			"date": p.Date, "type": p.Type, "amount": p.Amount,
			"category": p.Category, "note": p.Note, "counterparty": p.Counterparty,
		}
	}
	schema = domain.ResolveEntrySchema(schema)
	if err := domain.ValidateEntryData(schema, data); err != nil {
		return domain.EntryPayload{}, err
	}
	return domain.EntryPayload{
		SchemaID: schema.TemplateID,
		Data:     data,
	}, nil
}

// BuildTemplate creates xlsx bytes from schema field labels.
func BuildTemplate(schema domain.EntrySchema) ([]byte, error) {
	schema = domain.ResolveEntrySchema(schema)
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"
	for i, fdef := range schema.Fields {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, fdef.Label)
	}
	// example row for default template
	if schema.TemplateID == domain.TemplateDefault {
		ex := []string{"", "张三", "128.50", "2026-05-22", "午餐"}
		for ci, v := range ex {
			if ci >= len(schema.Fields) {
				break
			}
			cell, _ := excelize.CoordinatesToCellName(ci+1, 2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	} else if schema.TemplateID == domain.TemplateClassic {
		ex := []string{"2026-05-22", "支出", "128.50", "餐饮", "午餐", "餐厅A"}
		for ci, v := range ex {
			if ci >= len(schema.Fields) {
				break
			}
			cell, _ := excelize.CoordinatesToCellName(ci+1, 2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fillLegacyFlat(p *RowPreview) {
	if p.Cells == nil {
		return
	}
	p.Date = p.Cells["date"]
	p.Type = p.Cells["type"]
	p.Amount = p.Cells["amount"]
	p.Category = p.Cells["category"]
	p.Note = p.Cells["note"]
	p.Counterparty = p.Cells["counterparty"]
}

func validateRow(schema domain.EntrySchema, p *RowPreview) error {
	if err := domain.ValidateEntryData(schema, p.Cells); err != nil {
		return err
	}
	return nil
}

func normHeader(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func rowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func normType(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "收入", "income", "in":
		return "income"
	case "支出", "expense", "out", "exp":
		return "expense"
	default:
		return s
	}
}
