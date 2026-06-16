package importfile

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
	"github.com/xuri/excelize/v2"
)

func readXLSXFirstSheet(data []byte) ([]string, [][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, nil, importxlsx.ErrEmptyFile
	}
	all, err := f.GetRows(sheet)
	if err != nil {
		return nil, nil, err
	}
	if len(all) < 2 {
		return nil, nil, importxlsx.ErrNoDataRows
	}
	headers := all[0]
	var rows [][]string
	for _, r := range all[1:] {
		if lineEmpty(r) {
			continue
		}
		rows = append(rows, r)
		if len(rows) >= 5000 {
			return nil, nil, importxlsx.ErrTooManyRows
		}
	}
	if len(rows) == 0 {
		return nil, nil, importxlsx.ErrNoDataRows
	}
	return headers, rows, nil
}

func lineEmpty(r []string) bool {
	for _, c := range r {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
