package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAutoImportMarkdownTable_BypassesLLM(t *testing.T) {
	var gotCSV string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/import/sheet-csv") {
			http.NotFound(w, r)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		gotCSV, _ = body["csv"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":      "created",
			"tableId":   "T1",
			"tableName": "采购明细",
			"import":    map[string]any{"imported": 9, "skipped": 0},
			"parsedRows": 9,
		})
	}))
	defer srv.Close()

	l := &LedgerHTTP{BaseURL: srv.URL, UserID: "u1", SourceText: sampleProcurementMD}
	summary, ok := autoImportMarkdownTable(t.Context(), l, "L1", sampleProcurementMD)
	if !ok {
		t.Fatal("expected auto-import to run")
	}
	if !strings.Contains(summary, "9") {
		t.Fatalf("summary missing row count: %s", summary)
	}
	lines := strings.Split(strings.TrimSpace(gotCSV), "\n")
	if len(lines) != 10 {
		t.Fatalf("csv lines=%d want 10\n%s", len(lines), gotCSV)
	}
}

func TestAutoImportSkipsFollowUpWithoutTable(t *testing.T) {
	l := &LedgerHTTP{BaseURL: "http://127.0.0.1:9", UserID: "u1", SourceText: "你好"}
	if _, ok := autoImportMarkdownTable(t.Context(), l, "L1", "你好"); ok {
		t.Fatal("follow-up without markdown table must not auto-import")
	}
	if _, ok := autoImportMarkdownTable(t.Context(), l, "L1", "为什么我看账本里是空白的"); ok {
		t.Fatal("question without table must not auto-import")
	}
}

func TestLatestUserContentIgnoresHistory(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: sampleProcurementMD},
		{Role: "assistant", Content: "表格导入完成"},
		{Role: "user", Content: "你好"},
	}
	if got := latestUserContent(msgs); got != "你好" {
		t.Fatalf("got %q", got)
	}
}
