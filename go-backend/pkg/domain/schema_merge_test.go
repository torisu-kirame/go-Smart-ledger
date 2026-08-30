package domain

import "testing"

func TestRelaxRequiredForAbsentColumns(t *testing.T) {
	schema := DefaultEntrySchema()
	present := map[string]bool{"amount": true, "note": true}
	got := RelaxRequiredForAbsentColumns(schema, present)
	req := map[string]bool{}
	for _, f := range got.Fields {
		req[f.Key] = f.Required
	}
	if !req["amount"] {
		t.Fatal("amount should stay required when column present")
	}
	if req["date"] || req["bookkeeper"] {
		t.Fatalf("absent columns must not be required: %+v", req)
	}
}

func TestMergeEntrySchema_AddsMissingFields(t *testing.T) {
	base := EntrySchema{
		TemplateID: TemplateCustom,
		Fields: []EntryFieldDef{
			{Key: "item_name", Label: "物品名称", Type: FieldText},
			{Key: "amount", Label: "金额(元)", Type: FieldNumber},
		},
	}
	extra := EntrySchema{
		TemplateID: "import",
		Fields: []EntryFieldDef{
			{Key: "col_1", Label: "物品名称", Type: FieldText},
			{Key: "col_2", Label: "规格型号", Type: FieldText},
			{Key: "col_3", Label: "金额(元)", Type: FieldNumber},
		},
	}
	merged, remap := MergeEntrySchema(base, extra)
	if len(merged.Fields) != 3 {
		t.Fatalf("fields=%d want 3 %+v", len(merged.Fields), merged.Fields)
	}
	if remap["col_1"] != "item_name" || remap["col_3"] != "amount" {
		t.Fatalf("remap=%v", remap)
	}
	if remap["col_2"] != "col_2" {
		t.Fatalf("new field remap=%v", remap["col_2"])
	}
}
