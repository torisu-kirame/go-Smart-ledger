package ai

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LedgerHTTP calls ledger-api with the end-user JWT.
type LedgerHTTP struct {
	BaseURL    string
	AuthHeader string
	UserID     string
	// SourceText is the current user turn (may include a full Markdown table).
	// Used when the model truncates tool-call markdown/csv arguments.
	SourceText string
	HTTPClient *http.Client
}

func (l *LedgerHTTP) client() *http.Client {
	if l.HTTPClient != nil {
		return l.HTTPClient
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (l *LedgerHTTP) headers() http.Header {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	if l.AuthHeader != "" {
		h.Set("Authorization", l.AuthHeader)
	}
	if l.UserID != "" {
		h.Set("X-User-Id", l.UserID)
	}
	return h
}

func (l *LedgerHTTP) Request(ctx context.Context, method, apiPath string, query map[string]any, body any) (any, error) {
	u := strings.TrimRight(l.BaseURL, "/") + apiPath
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			if v == nil {
				continue
			}
			q.Set(k, fmt.Sprint(v))
		}
		u += "?" + q.Encode()
	}
	var reader io.Reader
	m := strings.ToUpper(method)
	if body != nil && m != http.MethodGet && m != http.MethodDelete {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, m, u, reader)
	if err != nil {
		return nil, err
	}
	req.Header = l.headers()
	resp, err := l.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 2000))
	}
	if resp.StatusCode == 204 || len(data) == 0 {
		return map[string]any{"ok": true, "status": resp.StatusCode}, nil
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		text := string(data)
		if len(text) > 16000 {
			text = text[:16000] + "...(truncated)"
		}
		return map[string]any{"ok": true, "status": resp.StatusCode, "text": text}, nil
	}
	return out, nil
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		fnTool("create_ledger",
			"Create a new simple multi-table ledger. Prefer this over raw call_ledger_api.",
			map[string]any{
				"name":              map[string]any{"type": "string", "description": "Ledger display name"},
				"type":              map[string]any{"type": "string", "description": "private (default) | multi"},
				"multiTableEnabled": map[string]any{"type": "boolean", "description": "Default true for simple ledgers"},
			},
			[]string{"name"},
		),
		fnTool("create_ledger_sheet",
			"Enable multi-table if needed and create a Sheet with custom fields. Do NOT create a field labeled 序号.",
			map[string]any{
				"ledgerId": map[string]any{"type": "string"},
				"name":     map[string]any{"type": "string", "description": "Sheet name, e.g. 采购明细"},
				"fields": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"key":      map[string]any{"type": "string", "description": "English key e.g. item_name"},
							"label":    map[string]any{"type": "string", "description": "Display label e.g. 物品名称"},
							"type":     map[string]any{"type": "string", "description": "text|number|date|user"},
							"required": map[string]any{"type": "boolean"},
						},
						"required": []string{"key", "label"},
					},
				},
			},
			[]string{"name", "fields"},
		),
		fnTool("append_ledger_entry",
			"Append one entry to a sheet. Body is sent as {entry:{tableId,data,...}} (required by API).",
			map[string]any{
				"ledgerId":  map[string]any{"type": "string"},
				"tableId":   map[string]any{"type": "string", "description": "Required for multi-table ledgers"},
				"data":      map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": "Field key→value matching sheet schema"},
				"amount":    map[string]any{"type": "string"},
				"note":      map[string]any{"type": "string"},
				"category":  map[string]any{"type": "string"},
				"entryType": map[string]any{"type": "string", "description": "expense|income|transfer → maps to entry.type"},
				"date":      map[string]any{"type": "string", "description": "YYYY-MM-DD"},
			},
			[]string{},
		),
		fnTool("append_ledger_entries_batch",
			"Append multiple JSON rows (max 100). For Markdown/CSV tables prefer import_markdown_table or import_sheet_csv.",
			map[string]any{
				"ledgerId": map[string]any{"type": "string"},
				"tableId":  map[string]any{"type": "string"},
				"rows": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}},
				},
			},
			[]string{"tableId", "rows"},
		),
		fnTool("import_markdown_table",
			"PREFERRED for Markdown tables. Converts MD→CSV and calls POST /import/sheet-csv (batch). "+
				"Pass markdown OR omit it — the server will use the full table from the user's message if richer. "+
				"Omit tableId to create a sheet; set tableId to append. NEVER use append_ledger_entry for multi-row tables.",
			map[string]any{
				"ledgerId":  map[string]any{"type": "string"},
				"tableId":   map[string]any{"type": "string", "description": "Existing sheet id → append; empty → create sheet"},
				"sheetName": map[string]any{"type": "string", "description": "New sheet name when creating"},
				"markdown":  map[string]any{"type": "string", "description": "Optional; server prefers fuller table from user message if tool args are truncated"},
			},
			nil,
		),
		fnTool("import_sheet_csv",
			"Import CSV text via POST /api/v1/ledgers/{id}/import/sheet-csv. Empty tableId creates sheet; with tableId appends rows at bottom.",
			map[string]any{
				"ledgerId":  map[string]any{"type": "string"},
				"tableId":   map[string]any{"type": "string"},
				"sheetName": map[string]any{"type": "string"},
				"csv":       map[string]any{"type": "string", "description": "Full CSV including header row"},
			},
			[]string{"csv"},
		),
		fnTool("list_ledgers", "List ledgers the current user can access.", map[string]any{}, nil),
		fnTool("get_ledger_summary",
			"Get ledger metadata including sheets/tables and field schemas.",
			map[string]any{"ledgerId": map[string]any{"type": "string"}},
			nil,
		),
		fnTool("search_ledger_rag",
			"Export ledger events as text chunks for analysis.",
			map[string]any{
				"ledgerId": map[string]any{"type": "string"},
				"limit":    map[string]any{"type": "integer"},
				"query":    map[string]any{"type": "string"},
			},
			nil,
		),
		fnTool("call_ledger_api",
			"Call allowlisted REST. For POST .../entries MUST use body {\"entry\":{\"tableId\":\"...\",\"data\":{...}}} not a flat tableId. Prefer specialized tools when possible.",
			map[string]any{
				"method": map[string]any{"type": "string"},
				"path":   map[string]any{"type": "string"},
				"query":  map[string]any{"type": "object", "additionalProperties": true},
				"body":   map[string]any{"type": "object", "additionalProperties": true},
			},
			[]string{"method", "path"},
		),
	}
}

func fnTool(name, desc string, props map[string]any, required []string) map[string]any {
	params := map[string]any{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": true,
	}
	if required != nil {
		params["required"] = required
	}
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        name,
			"description": desc,
			"parameters":   params,
		},
	}
}

func (l *LedgerHTTP) invokeTool(ctx context.Context, name string, args map[string]any, boundLedgerID string) string {
	switch name {
	case "create_ledger":
		return l.toolCreateLedger(ctx, args)
	case "create_ledger_sheet":
		return l.toolCreateSheet(ctx, args, boundLedgerID)
	case "append_ledger_entry":
		if n := sourceTableRowCount(l.SourceText); n > 1 {
			return fmt.Sprintf(
				"error: user message contains a Markdown table with %d rows. "+
					"Call import_markdown_table once (markdown may be omitted; server uses the full user table). "+
					"Do NOT append_ledger_entry row-by-row.",
				n,
			)
		}
		return l.toolAppendEntry(ctx, args, boundLedgerID)
	case "append_ledger_entries_batch":
		if n := sourceTableRowCount(l.SourceText); n > 1 {
			return fmt.Sprintf(
				"error: user message contains a Markdown table with %d rows. "+
					"Call import_markdown_table once instead of building JSON rows (args get truncated).",
				n,
			)
		}
		return l.toolAppendBatch(ctx, args, boundLedgerID)
	case "import_markdown_table":
		return l.toolImportMarkdown(ctx, args, boundLedgerID)
	case "import_sheet_csv":
		return l.toolImportSheetCSV(ctx, args, boundLedgerID)
	case "call_ledger_api":
		if n := sourceTableRowCount(l.SourceText); n > 1 {
			p, _ := args["path"].(string)
			pl := strings.ToLower(p)
			if strings.Contains(pl, "/entries") {
				return fmt.Sprintf(
					"error: user message has a %d-row Markdown table. Use import_markdown_table (not call_ledger_api .../entries).",
					n,
				)
			}
		}
		return l.toolCallAPI(ctx, args, boundLedgerID)
	case "list_ledgers":
		out, err := l.Request(ctx, http.MethodGet, "/api/v1/ledgers", nil, nil)
		if err != nil {
			return "error: " + err.Error()
		}
		return mustJSON(out)
	case "get_ledger_summary":
		lid := resolveLedgerID(args, boundLedgerID)
		if lid == "" {
			return "error: ledgerId required"
		}
		out, err := l.Request(ctx, http.MethodGet, "/api/v1/ledgers/"+lid, nil, nil)
		if err != nil {
			return "error: " + err.Error()
		}
		return mustJSON(out)
	case "search_ledger_rag":
		lid := resolveLedgerID(args, boundLedgerID)
		if lid == "" {
			return "error: ledgerId required"
		}
		q := map[string]any{}
		if lim, ok := args["limit"]; ok {
			q["limit"] = lim
		}
		if qq, ok := args["query"].(string); ok && qq != "" {
			q["q"] = qq
		}
		out, err := l.Request(ctx, http.MethodGet, "/api/v1/ledgers/"+lid+"/rag-export", q, nil)
		if err != nil {
			return "error: " + err.Error()
		}
		return truncateJSON(out, 16000)
	default:
		return "error: unknown tool " + name
	}
}

func (l *LedgerHTTP) toolCreateLedger(ctx context.Context, args map[string]any) string {
	name, _ := args["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return "error: name required"
	}
	lt, _ := args["type"].(string)
	if lt != "multi" {
		lt = "private"
	}
	multi := true
	if v, ok := args["multiTableEnabled"].(bool); ok {
		multi = v
	}
	uid := l.UserID
	body := map[string]any{
		"type":              lt,
		"name":              name,
		"creatorId":         uid,
		"members":           []map[string]any{{"id": uid, "role": "owner"}},
		"bookkeepingMode":   "simple",
		"multiTableEnabled": multi,
		"entrySchema":       map[string]any{"templateId": "custom", "fields": []any{}},
		"approvalPolicy":    map[string]any{"enabled": false, "threshold": 1},
	}
	out, err := l.Request(ctx, http.MethodPost, "/api/v1/ledgers", nil, body)
	if err != nil {
		return "error: " + err.Error()
	}
	return mustJSON(map[string]any{"ok": true, "ledger": out})
}

func (l *LedgerHTTP) toolCreateSheet(ctx context.Context, args map[string]any, bound string) string {
	lid := resolveLedgerID(args, bound)
	if lid == "" {
		return "error: ledgerId required"
	}
	sheetName, _ := args["name"].(string)
	sheetName = strings.TrimSpace(sheetName)
	if sheetName == "" {
		return "error: name required"
	}
	rawFields, _ := args["fields"].([]any)
	fields := normalizeFieldDefs(rawFields)
	if len(fields) == 0 {
		return "error: at least one field required (skip 序号)"
	}
	// Ensure multi-table on
	if _, err := l.Request(ctx, http.MethodPatch, "/api/v1/ledgers/"+lid+"/multi-table", nil, map[string]any{"enabled": true}); err != nil {
		// continue — may already be enabled
		_ = err
	}
	body := map[string]any{
		"name": sheetName,
		"entrySchema": map[string]any{
			"templateId": "custom",
			"fields":     fields,
		},
	}
	out, err := l.Request(ctx, http.MethodPost, "/api/v1/ledgers/"+lid+"/tables", nil, body)
	if err != nil {
		return "error: " + err.Error()
	}
	return mustJSON(map[string]any{"ok": true, "ledgerId": lid, "sheet": out})
}

func (l *LedgerHTTP) toolAppendEntry(ctx context.Context, args map[string]any, bound string) string {
	lid := resolveLedgerID(args, bound)
	if lid == "" {
		return "error: ledgerId required"
	}
	entry := buildEntryObject(args, l.UserID)
	if tid, _ := entry["tableId"].(string); strings.TrimSpace(tid) == "" {
		// Hint: multi-table ledgers need tableId from create_ledger_sheet / get_ledger_summary
		_ = tid
	}
	body := map[string]any{"entry": entry}
	out, err := l.Request(ctx, http.MethodPost, "/api/v1/ledgers/"+lid+"/entries", nil, body)
	if err != nil {
		return "error: " + err.Error() + " | tip: POST body must be {\"entry\":{\"tableId\":\"...\",\"data\":{...}}} and tableId must exist"
	}
	return mustJSON(out)
}

func (l *LedgerHTTP) toolAppendBatch(ctx context.Context, args map[string]any, bound string) string {
	lid := resolveLedgerID(args, bound)
	if lid == "" {
		return "error: ledgerId required"
	}
	tid, _ := args["tableId"].(string)
	tid = strings.TrimSpace(tid)
	if tid == "" {
		return "error: tableId required"
	}
	rows, _ := args["rows"].([]any)
	if len(rows) == 0 {
		return "error: rows required"
	}
	if len(rows) > 100 {
		return "error: max 100 rows per call; split batches"
	}
	ok, fail := 0, 0
	var errors []string
	for i, row := range rows {
		rm, okMap := row.(map[string]any)
		if !okMap {
			fail++
			errors = append(errors, fmt.Sprintf("row %d: not object", i+1))
			continue
		}
		entryArgs := map[string]any{"tableId": tid, "data": rm}
		entry := buildEntryObject(entryArgs, l.UserID)
		_, err := l.Request(ctx, http.MethodPost, "/api/v1/ledgers/"+lid+"/entries", nil, map[string]any{"entry": entry})
		if err != nil {
			fail++
			errors = append(errors, fmt.Sprintf("row %d: %s", i+1, err.Error()))
			continue
		}
		ok++
	}
	return mustJSON(map[string]any{"ok": fail == 0, "written": ok, "failed": fail, "errors": errors})
}

func (l *LedgerHTTP) toolImportMarkdown(ctx context.Context, args map[string]any, bound string) string {
	md, _ := args["markdown"].(string)
	md = preferRicherMarkdown(md, l.SourceText)
	if strings.TrimSpace(md) == "" {
		return "error: no markdown table found in tool args or user message"
	}
	_, rows, err := parseMarkdownTable(md)
	if err != nil {
		return "error: " + err.Error()
	}
	csvText, err := markdownToCSV(md)
	if err != nil {
		return "error: " + err.Error()
	}
	args2 := map[string]any{
		"csv":       csvText,
		"ledgerId":  args["ledgerId"],
		"tableId":   args["tableId"],
		"sheetName": args["sheetName"],
	}
	out := l.toolImportSheetCSV(ctx, args2, bound)
	// Annotate expected row count so the model cannot claim success with 1 row.
	return mustJSON(map[string]any{
		"expectedRows": len(rows),
		"importResult": jsonRawOrString(out),
		"note":         "expectedRows is the parsed Markdown row count; import.imported must match. If truncated tool args were fixed from user message, trust expectedRows.",
	})
}

func jsonRawOrString(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err == nil {
		return v
	}
	return s
}

func (l *LedgerHTTP) toolImportSheetCSV(ctx context.Context, args map[string]any, bound string) string {
	lid := resolveLedgerID(args, bound)
	if lid == "" {
		return "error: ledgerId required"
	}
	csvText, _ := args["csv"].(string)
	csvText = strings.TrimSpace(csvText)
	// If CSV looks too short vs SourceText markdown, rebuild from source table
	if l.SourceText != "" {
		if md := preferRicherMarkdown("", l.SourceText); md != "" {
			_, srcRows, err := parseMarkdownTable(md)
			if err == nil && len(srcRows) > 0 {
				csvLines := 0
				if csvText != "" {
					csvLines = len(strings.Split(strings.TrimSpace(csvText), "\n")) - 1
					if csvLines < 0 {
						csvLines = 0
					}
				}
				if csvLines < len(srcRows) {
					if rebuilt, err := markdownToCSV(md); err == nil {
						csvText = rebuilt
					}
				}
			}
		}
	}
	if csvText == "" {
		return "error: csv required"
	}
	tid, _ := args["tableId"].(string)
	sheetName, _ := args["sheetName"].(string)
	if strings.TrimSpace(sheetName) == "" {
		sheetName = "导入明细"
	}
	body := map[string]any{
		"csv":       csvText,
		"tableId":   strings.TrimSpace(tid),
		"sheetName": sheetName,
		"signerId":  l.UserID,
	}
	out, err := l.Request(ctx, http.MethodPost, "/api/v1/ledgers/"+lid+"/import/sheet-csv", nil, body)
	if err != nil {
		return "error: " + err.Error()
	}
	return mustJSON(map[string]any{
		"ok":          true,
		"ledgerId":    lid,
		"result":      out,
		"instruction": "Tell user mode(created|appended), tableId, and import.imported count. Do not append rows one-by-one.",
	})
}

func markdownToCSV(md string) (string, error) {
	headers, rows, err := parseMarkdownTable(md)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	w.UseCRLF = false
	if err := w.Write(headers); err != nil {
		return "", err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (l *LedgerHTTP) toolCallAPI(ctx context.Context, args map[string]any, bound string) string {
	method, _ := args["method"].(string)
	p, _ := args["path"].(string)
	norm, err := normalizeAPIPath(p, bound)
	if err != nil {
		return "error: " + err.Error()
	}
	if err := assertAPIAllowed(method, norm); err != nil {
		return "error: " + err.Error()
	}
	var query map[string]any
	if q, ok := args["query"].(map[string]any); ok {
		query = q
	}
	var body any
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "POST" || m == "PATCH" || m == "PUT" {
		b, _ := args["body"].(map[string]any)
		if b == nil {
			b = map[string]any{}
		}
		if strings.HasSuffix(norm, "/entries") || strings.Contains(norm, "/entries/") {
			b = wrapEntriesBody(b, l.UserID)
		}
		body = b
	}
	out, err := l.Request(ctx, m, norm, query, body)
	if err != nil {
		return "error: " + err.Error()
	}
	return mustJSON(map[string]any{"ok": true, "method": m, "path": norm, "data": out})
}

func wrapEntriesBody(body map[string]any, userID string) map[string]any {
	if entry, ok := body["entry"].(map[string]any); ok {
		if sid, _ := entry["signerId"].(string); strings.TrimSpace(sid) == "" {
			entry["signerId"] = userID
		}
		body["entry"] = entry
		return body
	}
	return map[string]any{"entry": buildEntryObject(body, userID)}
}

func buildEntryObject(args map[string]any, userID string) map[string]any {
	entry := map[string]any{"signerId": userID}
	if tid, ok := args["tableId"].(string); ok && strings.TrimSpace(tid) != "" {
		entry["tableId"] = strings.TrimSpace(tid)
	}
	if sid, ok := args["schemaId"].(string); ok && strings.TrimSpace(sid) != "" {
		entry["schemaId"] = strings.TrimSpace(sid)
	}
	if data := stringifyMap(args["data"]); len(data) > 0 {
		entry["data"] = data
	}
	if v, ok := args["date"].(string); ok && v != "" {
		entry["date"] = v
	}
	if v, ok := args["amount"].(string); ok && v != "" {
		entry["amount"] = v
	}
	if v, ok := args["note"].(string); ok && v != "" {
		entry["note"] = v
	}
	if v, ok := args["category"].(string); ok && v != "" {
		entry["category"] = v
	}
	if v, ok := args["type"].(string); ok && v != "" {
		entry["type"] = v
	} else if v, ok := args["entryType"].(string); ok && v != "" {
		entry["type"] = v
	}
	if sid, ok := args["signerId"].(string); ok && strings.TrimSpace(sid) != "" {
		entry["signerId"] = strings.TrimSpace(sid)
	}
	return entry
}

func normalizeFieldDefs(raw []any) []map[string]any {
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		f, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := strings.TrimSpace(fmt.Sprint(f["key"]))
		label := strings.TrimSpace(fmt.Sprint(f["label"]))
		if key == "" || label == "" {
			continue
		}
		if label == "序号" || label == "编号" || label == "#" || key == "seq" || key == "no" {
			continue
		}
		typ := strings.TrimSpace(fmt.Sprint(f["type"]))
		if typ == "" || typ == "<nil>" {
			typ = "text"
		}
		if typ != "text" && typ != "number" && typ != "date" && typ != "user" {
			typ = "text"
		}
		req := false
		switch v := f["required"].(type) {
		case bool:
			req = v
		case string:
			req = v == "true" || v == "1"
		}
		out = append(out, map[string]any{
			"key": key, "label": label, "type": typ, "required": req,
		})
	}
	return out
}

func stringifyMap(v any) map[string]string {
	m, ok := v.(map[string]any)
	if !ok || len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, val := range m {
		if val == nil {
			continue
		}
		out[k] = fmt.Sprint(val)
	}
	return out
}

func resolveLedgerID(args map[string]any, bound string) string {
	if v, ok := args["ledgerId"].(string); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(bound)
}

func sourceTableRowCount(source string) int {
	md := preferRicherMarkdown("", source)
	if md == "" {
		return 0
	}
	_, rows, err := parseMarkdownTable(md)
	if err != nil {
		return 0
	}
	return len(rows)
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	s := string(b)
	if len(s) > 16000 {
		return s[:16000] + "...(truncated)"
	}
	return s
}

func truncateJSON(v any, n int) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "error: " + err.Error()
	}
	s := string(b)
	if len(s) > n {
		return s[:n] + "...(truncated)"
	}
	return s
}
