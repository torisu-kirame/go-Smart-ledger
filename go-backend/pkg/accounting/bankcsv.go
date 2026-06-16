package accounting

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
)

// ParseBankCSV reads common bank export CSV (date, description, amount columns).
func ParseBankCSV(r io.Reader, filename string) (*BankStatement, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("csv: no data rows")
	}
	header := rows[0]
	dateCol, descCol, amtCol := detectColumns(header)
	if dateCol < 0 || amtCol < 0 {
		return nil, fmt.Errorf("csv: need date and amount columns")
	}
	id, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	stmt := &BankStatement{
		ID:       id,
		Filename: filename,
		Lines:    make([]BankStatementLine, 0, len(rows)-1),
	}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		date := cell(row, dateCol)
		desc := ""
		if descCol >= 0 {
			desc = cell(row, descCol)
		}
		amt := cell(row, amtCol)
		if date == "" || amt == "" {
			continue
		}
		lid, _ := snowflake.NextString()
		stmt.Lines = append(stmt.Lines, BankStatementLine{
			ID:          lid,
			Date:        normalizeDate(date),
			Description: desc,
			Amount:      strings.ReplaceAll(amt, ",", ""),
		})
	}
	if len(stmt.Lines) == 0 {
		return nil, fmt.Errorf("csv: no valid lines")
	}
	return stmt, nil
}

func detectColumns(header []string) (dateCol, descCol, amtCol int) {
	dateCol, descCol, amtCol = -1, -1, -1
	for i, h := range header {
		lower := strings.ToLower(strings.TrimSpace(h))
		switch {
		case strings.Contains(lower, "日期") || lower == "date" || strings.Contains(lower, "交易时间"):
			dateCol = i
		case strings.Contains(lower, "摘要") || strings.Contains(lower, "说明") || lower == "description" || strings.Contains(lower, "备注"):
			if descCol < 0 {
				descCol = i
			}
		case strings.Contains(lower, "金额") || lower == "amount" || strings.Contains(lower, "收入") || strings.Contains(lower, "支出"):
			if amtCol < 0 {
				amtCol = i
			}
		}
	}
	return
}

func cell(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
