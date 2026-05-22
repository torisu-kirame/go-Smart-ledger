package importxlsx

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
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

// RowPreview is one parsed row with optional validation error.
type RowPreview struct {
	Line    int    `json:"line"`
	Date    string `json:"date"`
	Type    string `json:"type"`
	Amount  string `json:"amount"`
	Category string `json:"category,omitempty"`
	Note    string `json:"note,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Parse reads xlsx bytes and returns preview rows.
func Parse(data []byte) ([]RowPreview, error) {
	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, ErrEmptyFile
	}
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
	required := []string{"日期", "类型", "金额"}
	for _, k := range required {
		if _, ok := header[k]; !ok {
			return nil, fmt.Errorf("missing column: %s", k)
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
		p := RowPreview{
			Line: line + 1,
			Date: cell(r, header["日期"]),
			Type: normType(cell(r, header["类型"])),
			Amount: strings.TrimSpace(cell(r, header["金额"])),
		}
		if idx, ok := header["分类"]; ok {
			p.Category = cell(r, idx)
		}
		if idx, ok := header["备注"]; ok {
			p.Note = cell(r, idx)
		}
		if idx, ok := header["对方"]; ok {
			p.Counterparty = cell(r, idx)
		}
		if err := validateRow(&p); err != nil {
			p.Error = err.Error()
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, ErrNoDataRows
	}
	return out, nil
}

// ToEntry converts valid preview row to domain entry.
func ToEntry(p RowPreview) (domain.EntryPayload, error) {
	if p.Error != "" {
		return domain.EntryPayload{}, errors.New(p.Error)
	}
	return domain.EntryPayload{
		Date:         p.Date,
		Type:         p.Type,
		Amount:       p.Amount,
		Category:     p.Category,
		Note:         p.Note,
		Counterparty: p.Counterparty,
	}, nil
}

// BuildTemplate creates an xlsx template file bytes.
func BuildTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"
	headers := []string{"日期", "类型", "金额", "分类", "备注", "对方"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	examples := [][]string{
		{"2026-05-22", "支出", "128.50", "餐饮", "午餐", "餐厅A"},
		{"2026-05-22", "收入", "5000.00", "工资", "五月工资", "公司"},
	}
	for ri, row := range examples {
		for ci, v := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

func validateRow(p *RowPreview) error {
	if p.Date == "" {
		return errors.New("日期不能为空")
	}
	if p.Type != "income" && p.Type != "expense" {
		return errors.New("类型须为 收入/支出 或 income/expense")
	}
	if p.Amount == "" {
		return errors.New("金额不能为空")
	}
	if _, err := strconv.ParseFloat(strings.ReplaceAll(p.Amount, ",", ""), 64); err != nil {
		return errors.New("金额格式无效")
	}
	return nil
}
