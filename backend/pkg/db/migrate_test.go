package db

import "testing"

func TestSplitDSN(t *testing.T) {
	dbName, admin, err := splitDSN("root:123456@tcp(127.0.0.1:3306)/smart_ledger?parseTime=true&charset=utf8mb4")
	if err != nil {
		t.Fatal(err)
	}
	if dbName != "smart_ledger" {
		t.Fatalf("dbName=%q", dbName)
	}
	if admin == "" || containsDSNDB(admin, "smart_ledger") {
		t.Fatalf("adminDSN should omit database: %q", admin)
	}
}

func containsDSNDB(dsn, db string) bool {
	return len(dsn) > 0 && (dsn == "/"+db || len(dsn) > len(db) && dsn[len(dsn)-len(db)-1:] == "/"+db)
}
