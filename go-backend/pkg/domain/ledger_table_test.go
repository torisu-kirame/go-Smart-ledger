package domain

import "testing"

func TestNormalizeLedgerTablesDefault(t *testing.T) {
	meta := &LedgerMeta{EntrySchema: DefaultEntrySchema()}
	NormalizeLedgerTables(meta)
	if len(meta.Tables) != 1 || meta.Tables[0].ID != DefaultTableID {
		t.Fatalf("expected default table, got %+v", meta.Tables)
	}
}

func TestNormalizeLedgerTablesBlankMulti(t *testing.T) {
	meta := &LedgerMeta{
		MultiTableEnabled: true,
		EntrySchema:       EntrySchema{TemplateID: TemplateCustom},
		Tables:            nil,
	}
	NormalizeLedgerTables(meta)
	if len(meta.Tables) != 0 {
		t.Fatalf("expected empty tables, got %+v", meta.Tables)
	}
	if meta.EntrySchema.TemplateID != TemplateCustom {
		t.Fatalf("expected custom template id")
	}
}

func TestSchemaForTableMulti(t *testing.T) {
	meta := &LedgerMeta{
		MultiTableEnabled: true,
		Tables: []LedgerTable{
			{ID: "a", Name: "A", EntrySchema: ClassicEntrySchema()},
		},
	}
	sch, err := SchemaForTable(meta, "a")
	if err != nil || sch.TemplateID != TemplateClassic {
		t.Fatalf("schema: %v err %v", sch, err)
	}
}
