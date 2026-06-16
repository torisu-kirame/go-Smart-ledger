package domain

import "testing"

func TestNormalizeLedgerTablesDefault(t *testing.T) {
	meta := &LedgerMeta{EntrySchema: DefaultEntrySchema()}
	NormalizeLedgerTables(meta)
	if len(meta.Tables) != 1 || meta.Tables[0].ID != DefaultTableID {
		t.Fatalf("expected default table, got %+v", meta.Tables)
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
