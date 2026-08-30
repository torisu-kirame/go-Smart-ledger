package importfile

import (
	"strings"
	"testing"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

func TestPrepareCSVForNewSheet_SkipsSeq(t *testing.T) {
	csv := "序号,物品名称,金额(元),备注\n" +
		"1,CPU,\"2,050.00\",办公\n" +
		"2,主板,1099.00,办公\n"
	schema, rows, err := PrepareCSVForNewSheet([]byte(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Fields) != 3 {
		t.Fatalf("fields=%d want 3: %+v", len(schema.Fields), schema.Fields)
	}
	for _, f := range schema.Fields {
		if strings.Contains(f.Label, "序号") {
			t.Fatal("序号 should be skipped")
		}
		if f.Required {
			t.Fatalf("import fields should not be required: %s", f.Label)
		}
	}
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Cells[schema.Fields[0].Key] != "CPU" {
		t.Fatalf("cells=%v", rows[0].Cells)
	}
}

func TestMapRowsByHeaders_Append(t *testing.T) {
	schema := domain.EntrySchema{
		TemplateID: "import",
		Fields: []domain.EntryFieldDef{
			{Key: "item_name", Label: "物品名称", Type: domain.FieldText},
			{Key: "amount", Label: "金额(元)", Type: domain.FieldNumber},
		},
	}
	headers := []string{"序号", "物品名称", "金额(元)"}
	raw := [][]string{{"3", "内存", "499.00"}}
	rows := MapRowsByHeaders(headers, raw, schema)
	if len(rows) != 1 {
		t.Fatal(rows)
	}
	if rows[0].Error != "" {
		t.Fatal(rows[0].Error)
	}
	if rows[0].Cells["item_name"] != "内存" || rows[0].Cells["amount"] != "499.00" {
		t.Fatalf("%v", rows[0].Cells)
	}
}
