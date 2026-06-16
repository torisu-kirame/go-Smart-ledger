package accounting

import "testing"

func TestValidateJournal_balanced(t *testing.T) {
	chart := DefaultChart("1")
	j := &JournalEntry{
		Date: "2026-05-01",
		Lines: []JournalLine{
			{AccountCode: "1002", Debit: "1000.00"},
			{AccountCode: "6001", Credit: "1000.00"},
		},
	}
	if err := ValidateJournal(j, chart); err != nil {
		t.Fatal(err)
	}
}

func TestValidateJournal_unbalanced(t *testing.T) {
	chart := DefaultChart("1")
	j := &JournalEntry{
		Date: "2026-05-01",
		Lines: []JournalLine{
			{AccountCode: "1002", Debit: "1000"},
			{AccountCode: "6001", Credit: "900"},
		},
	}
	if err := ValidateJournal(j, chart); err != ErrUnbalanced {
		t.Fatalf("want ErrUnbalanced, got %v", err)
	}
}
