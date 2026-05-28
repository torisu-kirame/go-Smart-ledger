package importfile

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
)

func readCSVRows(data []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	var all [][]string
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("csv: %w", err)
		}
		all = append(all, rec)
	}
	if len(all) < 2 {
		return nil, nil, importxlsx.ErrNoDataRows
	}
	headers := all[0]
	var rows [][]string
	for _, r := range all[1:] {
		if rowEmpty(r) {
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

func rowEmpty(r []string) bool {
	for _, c := range r {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
