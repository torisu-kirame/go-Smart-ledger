package importfile

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
)

// ParseWithSchema parses xlsx or csv using an existing schema.
func ParseWithSchema(data []byte, filename, sheetName string, schema domain.EntrySchema) ([]importxlsx.RowPreview, error) {
	if IsCSVName(filename) {
		headers, rows, err := readCSVRows(data)
		if err != nil {
			return nil, err
		}
		_ = headers
		return importxlsx.ParseRowsWithSchema(schema, rows, 1)
	}
	return importxlsx.ParseSheet(data, sheetName, schema)
}
