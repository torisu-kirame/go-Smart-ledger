package aiproxy

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
)

type mockLedgerBackend struct {
	ledgers map[string]*domain.LedgerMeta
	rag     map[string]*ledgersvc.RAGExport
	verify  map[string]bool
}

func (m *mockLedgerBackend) ListForUser(_ context.Context, _ string) ([]*domain.LedgerMeta, error) {
	out := make([]*domain.LedgerMeta, 0, len(m.ledgers))
	for _, meta := range m.ledgers {
		out = append(out, meta)
	}
	return out, nil
}

func (m *mockLedgerBackend) GetForUser(_ context.Context, ledgerID, _ string) (*domain.LedgerMeta, error) {
	meta, ok := m.ledgers[ledgerID]
	if !ok {
		return nil, domain.ErrLedgerNotFound
	}
	return meta, nil
}

func (m *mockLedgerBackend) ExportRAG(_ context.Context, ledgerID, _ string) (*ledgersvc.RAGExport, error) {
	if exp, ok := m.rag[ledgerID]; ok {
		return exp, nil
	}
	return nil, domain.ErrLedgerNotFound
}

func (m *mockLedgerBackend) Verify(_ context.Context, ledgerID string) (bool, error) {
	if v, ok := m.verify[ledgerID]; ok {
		return v, nil
	}
	return false, domain.ErrLedgerNotFound
}

func TestListLedgersTool(t *testing.T) {
	backend := &mockLedgerBackend{
		ledgers: map[string]*domain.LedgerMeta{
			"L1": {ID: "L1", Name: "测试账本", Type: domain.LedgerPrivate, LatestSeq: 3, AnchorStatus: "none"},
		},
	}
	tool := listLedgersTool{&ledgerToolCtx{svc: backend, userID: "u1"}}
	out, err := tool.Call(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["id"] != "L1" {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestGetLedgerSummaryUsesBoundLedger(t *testing.T) {
	now := time.Now().UTC()
	backend := &mockLedgerBackend{
		ledgers: map[string]*domain.LedgerMeta{
			"bound-id": {ID: "bound-id", Name: "Bound", Type: domain.LedgerMulti, LatestSeq: 9, LatestRoot: "abc", UpdatedAt: now},
		},
	}
	tool := getLedgerSummaryTool{&ledgerToolCtx{svc: backend, userID: "u1", defaultLedgerID: "bound-id"}}
	out, err := tool.Call(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	var summary map[string]any
	if err := json.Unmarshal([]byte(out), &summary); err != nil {
		t.Fatal(err)
	}
	if summary["id"] != "bound-id" || summary["latestSeq"].(float64) != 9 {
		t.Fatalf("unexpected summary: %s", out)
	}
}

func TestSearchLedgerRAGFiltersQuery(t *testing.T) {
	backend := &mockLedgerBackend{
		ledgers: map[string]*domain.LedgerMeta{
			"L1": {ID: "L1", Name: "RAG", LatestSeq: 2},
		},
		rag: map[string]*ledgersvc.RAGExport{
			"L1": {
				LedgerID: "L1", LedgerName: "RAG",
				Chunks: []ledgersvc.RAGChunk{
					{Seq: 1, Type: "entry", Text: "午餐报销 120 元"},
					{Seq: 2, Type: "entry", Text: "收到货款 5000 元"},
				},
			},
		},
	}
	tool := searchLedgerRAGTool{&ledgerToolCtx{svc: backend, userID: "u1", defaultLedgerID: "L1"}}
	out, err := tool.Call(context.Background(), `{"query":"报销","limit":10}`)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out, "午餐报销") || contains(out, "收到货款") {
		t.Fatalf("filter failed: %s", out)
	}
}

func TestVerifyLedgerTool(t *testing.T) {
	backend := &mockLedgerBackend{
		ledgers: map[string]*domain.LedgerMeta{"L1": {ID: "L1", Name: "V"}},
		verify:  map[string]bool{"L1": true},
	}
	tool := verifyLedgerTool{&ledgerToolCtx{svc: backend, userID: "u1"}}
	out, err := tool.Call(context.Background(), `{"ledgerId":"L1"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(out, `"valid":true`) {
		t.Fatalf("expected valid=true: %s", out)
	}
}

func TestBuildAgentInput(t *testing.T) {
	in := buildAgentInput([]ChatMessage{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "您好"},
		{Role: "user", Content: "列出账本"},
	})
	if !containsAll(in, "Current question:", "列出账本", "User: 你好") {
		t.Fatalf("unexpected agent input: %q", in)
	}
}

func TestBuildAgentSystemPrefixBoundLedger(t *testing.T) {
	prefix := buildAgentSystemPrefix(nil, "ledger-42")
	if !contains(prefix, "ledger-42") || !contains(prefix, "list_ledgers") {
		t.Fatalf("missing bound ledger hint: %q", prefix[:min(200, len(prefix))])
	}
}
