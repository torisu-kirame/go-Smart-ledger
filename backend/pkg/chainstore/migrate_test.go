package chainstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestSortMigrateRowsMetaFirst(t *testing.T) {
	rows := []StateRow{
		{Key: "smartledger:ledger:a:event:1", Value: json.RawMessage(`{}`)},
		{Key: "smartledger:ledger:b", Value: json.RawMessage(`{}`)},
		{Key: "smartledger:ledger:a", Value: json.RawMessage(`{}`)},
	}
	sortMigrateRows(rows)
	if rows[0].Key != "smartledger:ledger:a" && rows[0].Key != "smartledger:ledger:b" {
		t.Fatalf("meta not first: %v", rows[0].Key)
	}
	if !isLedgerMetaKey(rows[0].Key) {
		t.Fatalf("first row should be meta: %s", rows[0].Key)
	}
}

func TestShouldSkipMigrateKey(t *testing.T) {
	if !shouldSkipMigrateKey("smartledger:ledger:x:__keys__") {
		t.Fatal("internal index key")
	}
	if shouldSkipMigrateKey("smartledger:ledger:x:event:1") {
		t.Fatal("event key should migrate")
	}
}

func TestStateValuesEqual(t *testing.T) {
	a := json.RawMessage(`{"a":1}`)
	b := json.RawMessage(`{ "a" : 1 }`)
	if !stateValuesEqual(a, b) {
		t.Fatal("semantic equal")
	}
}

type mapStore struct {
	backend Backend
	data    map[string]json.RawMessage
}

func newMapStore(backend Backend) *mapStore {
	return &mapStore{backend: backend, data: make(map[string]json.RawMessage)}
}

func (m *mapStore) Backend() Backend { return m.backend }
func (m *mapStore) Ping(context.Context) error { return nil }
func (m *mapStore) Status(context.Context) (*Status, error) {
	return &Status{Backend: string(m.backend)}, nil
}
func (m *mapStore) Submit(_ context.Context, tx TxRequest) error {
	if tx.Key == "" {
		return fmt.Errorf("empty key")
	}
	m.data[tx.Key] = append(json.RawMessage(nil), tx.Value...)
	return nil
}
func (m *mapStore) Query(_ context.Context, sql string, params ...interface{}) ([]StateRow, error) {
	if len(params) == 1 {
		if prefix, ok := params[0].(string); ok && strings.Contains(sql, "LIKE") {
			var rows []StateRow
			for k, v := range m.data {
				if strings.HasPrefix(k, strings.TrimSuffix(prefix, "%")) {
					rows = append(rows, StateRow{Key: k, Value: append(json.RawMessage(nil), v...)})
				}
			}
			return rows, nil
		}
		if key, ok := params[0].(string); ok && strings.Contains(sql, "key = ?") {
			if v, ok := m.data[key]; ok {
				return []StateRow{{Key: key, Value: append(json.RawMessage(nil), v...)}}, nil
			}
			return nil, nil
		}
	}
	return nil, fmt.Errorf("unsupported query: %s", sql)
}
func (m *mapStore) GetRaw(context.Context, string) ([]byte, error) { return nil, nil }
func (m *mapStore) BaseURL() string { return "" }

func TestMigrateMiniLedgerToFISCO_DryRun(t *testing.T) {
	src := newMapStore(BackendMiniLedger)
	src.data["smartledger:ledger:a"] = json.RawMessage(`{"name":"a"}`)
	src.data["smartledger:ledger:a:__keys__"] = json.RawMessage(`[]`)
	src.data["smartledger:ledger:a:event:1"] = json.RawMessage(`{"v":1}`)

	res, err := MigrateMiniLedgerToFISCO(context.Background(), src, nil, MigrateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 2 {
		t.Fatalf("total=%d want 2 (skip __keys__)", res.Total)
	}
	if res.Written != 2 {
		t.Fatalf("written=%d want 2", res.Written)
	}
}

func TestMigrateMiniLedgerToFISCO_Verify(t *testing.T) {
	src := newMapStore(BackendMiniLedger)
	dst := newMapStore(BackendFISCO)
	src.data["smartledger:ledger:b"] = json.RawMessage(`{"name":"b"}`)
	src.data["smartledger:ledger:b:event:1"] = json.RawMessage(`{"n":1}`)

	res, err := MigrateMiniLedgerToFISCO(context.Background(), src, dst, MigrateOptions{Verify: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != 2 || res.Verified != 2 || res.Failed != 0 {
		t.Fatalf("result=%+v", res)
	}
	if len(dst.data) != 2 {
		t.Fatalf("dst keys=%d want 2", len(dst.data))
	}
}
