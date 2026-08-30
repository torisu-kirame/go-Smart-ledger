package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToolImportMarkdown_UsesSourceWhenToolArgsTruncated(t *testing.T) {
	var gotCSV string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ledgers/L1/import/sheet-csv" {
			http.NotFound(w, r)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		gotCSV, _ = gotBody["csv"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"import": map[string]any{
				"mode": "created",
				"import": map[string]any{"imported": 9, "skipped": 0},
				"parsedRows": 9,
			},
		})
	}))
	defer srv.Close()

	truncatedToolMD := `| 序号 | 物品名称 | 备注 |
| :--- | :--- | :--- |
| 9 | 键鼠套装 | 技术岗办公用 |
`
	l := &LedgerHTTP{
		BaseURL:    srv.URL,
		UserID:     "u1",
		SourceText: sampleProcurementMD,
	}
	out := l.toolImportMarkdown(t.Context(), map[string]any{
		"markdown":  truncatedToolMD,
		"sheetName": "采购测试",
	}, "L1")
	if strings.HasPrefix(out, "error:") {
		t.Fatal(out)
	}
	lines := strings.Split(strings.TrimSpace(gotCSV), "\n")
	if len(lines) != 10 {
		t.Fatalf("csv data lines=%d want 10 (hdr+9)\ncsv=%s\nout=%s", len(lines), gotCSV, out)
	}
	if !strings.Contains(gotCSV, "CPU") || !strings.Contains(gotCSV, "键鼠套装") {
		t.Fatalf("csv missing rows: %s", gotCSV)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatal(err)
	}
	if int(parsed["expectedRows"].(float64)) != 9 {
		t.Fatalf("expectedRows=%v out=%s", parsed["expectedRows"], out)
	}
}
