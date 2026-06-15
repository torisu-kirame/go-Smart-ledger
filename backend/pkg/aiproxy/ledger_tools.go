package aiproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	lctools "github.com/tmc/langchaingo/tools"
)

// LedgerToolBackend is the subset of ledgersvc used by agent tools.
type LedgerToolBackend interface {
	ListForUser(ctx context.Context, userID string) ([]*domain.LedgerMeta, error)
	GetForUser(ctx context.Context, ledgerID, userID string) (*domain.LedgerMeta, error)
	ExportRAG(ctx context.Context, ledgerID, userID string) (*ledgersvc.RAGExport, error)
	Verify(ctx context.Context, ledgerID string) (bool, error)
}

type ledgerToolCtx struct {
	svc             LedgerToolBackend
	userID          string
	defaultLedgerID string
}

// NewLedgerTools builds MRKL tools scoped to the authenticated user.
func NewLedgerTools(svc LedgerToolBackend, userID, defaultLedgerID string) []lctools.Tool {
	ctx := &ledgerToolCtx{
		svc:             svc,
		userID:          strings.TrimSpace(userID),
		defaultLedgerID: strings.TrimSpace(defaultLedgerID),
	}
	return []lctools.Tool{
		listLedgersTool{ctx},
		getLedgerSummaryTool{ctx},
		searchLedgerRAGTool{ctx},
		verifyLedgerTool{ctx},
	}
}

type listLedgersTool struct{ *ledgerToolCtx }

func (t listLedgersTool) Name() string { return "list_ledgers" }

func (t listLedgersTool) Description() string {
	return `List ledgers the current user can access. Input is optional JSON {} or empty. Returns JSON array with id, name, type, latestSeq, anchorStatus.`
}

func (t listLedgersTool) Call(ctx context.Context, input string) (string, error) {
	if t.svc == nil {
		return "ledger service unavailable", nil
	}
	items, err := t.svc.ListForUser(ctx, t.userID)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	type row struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Type         string `json:"type"`
		LatestSeq    uint64 `json:"latestSeq"`
		AnchorStatus string `json:"anchorStatus"`
	}
	out := make([]row, 0, len(items))
	for _, m := range items {
		out = append(out, row{
			ID:           m.ID,
			Name:         m.Name,
			Type:         string(m.Type),
			LatestSeq:    m.LatestSeq,
			AnchorStatus: m.AnchorStatus,
		})
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

type getLedgerSummaryTool struct{ *ledgerToolCtx }

func (t getLedgerSummaryTool) Name() string { return "get_ledger_summary" }

func (t getLedgerSummaryTool) Description() string {
	return `Get metadata for one ledger. Input JSON: {"ledgerId":"..."} or plain ledger id string. Uses bound ledger when ledgerId omitted.`
}

func (t getLedgerSummaryTool) Call(ctx context.Context, input string) (string, error) {
	if t.svc == nil {
		return "ledger service unavailable", nil
	}
	ledgerID, err := t.resolveLedgerID(input)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	meta, err := t.svc.GetForUser(ctx, ledgerID, t.userID)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	summary := map[string]any{
		"id":           meta.ID,
		"name":         meta.Name,
		"type":         meta.Type,
		"latestSeq":    meta.LatestSeq,
		"latestRoot":   meta.LatestRoot,
		"anchorStatus": meta.AnchorStatus,
		"memberCount":  len(meta.Members),
		"createdAt":    meta.CreatedAt,
		"updatedAt":    meta.UpdatedAt,
	}
	b, _ := json.Marshal(summary)
	return string(b), nil
}

type searchLedgerRAGTool struct{ *ledgerToolCtx }

func (t searchLedgerRAGTool) Name() string { return "search_ledger_rag" }

func (t searchLedgerRAGTool) Description() string {
	return `Export ledger events as text chunks for Q&A. Input JSON: {"ledgerId":"...","limit":40,"query":"optional keyword"}. Uses bound ledger when ledgerId omitted.`
}

func (t searchLedgerRAGTool) Call(ctx context.Context, input string) (string, error) {
	if t.svc == nil {
		return "ledger service unavailable", nil
	}
	var args struct {
		LedgerID string `json:"ledgerId"`
		Limit    int    `json:"limit"`
		Query    string `json:"query"`
	}
	_ = parseToolInput(input, &args)
	ledgerID := strings.TrimSpace(args.LedgerID)
	if ledgerID == "" {
		var err error
		ledgerID, err = t.resolveLedgerID(input)
		if err != nil {
			return fmt.Sprintf("error: %v", err), nil
		}
	}
	limit := args.Limit
	if limit <= 0 {
		limit = 40
	}
	if limit > 120 {
		limit = 120
	}
	export, err := t.svc.ExportRAG(ctx, ledgerID, t.userID)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	query := strings.ToLower(strings.TrimSpace(args.Query))
	chunks := export.Chunks
	if query != "" {
		filtered := make([]ledgersvc.RAGChunk, 0, len(chunks))
		for _, c := range chunks {
			if strings.Contains(strings.ToLower(c.Text), query) {
				filtered = append(filtered, c)
			}
		}
		chunks = filtered
	}
	if len(chunks) > limit {
		chunks = chunks[:limit]
	}
	type chunkRow struct {
		Seq  uint64 `json:"seq"`
		Type string `json:"type"`
		Text string `json:"text"`
	}
	rows := make([]chunkRow, 0, len(chunks))
	for _, c := range chunks {
		rows = append(rows, chunkRow{Seq: c.Seq, Type: c.Type, Text: c.Text})
	}
	out := map[string]any{
		"ledgerId":   export.LedgerID,
		"ledgerName": export.LedgerName,
		"total":      len(export.Chunks),
		"returned":   len(rows),
		"chunks":     rows,
	}
	b, _ := json.Marshal(out)
	s := string(b)
	const maxLen = 16000
	if len(s) > maxLen {
		s = s[:maxLen] + "...(truncated)"
	}
	return s, nil
}

type verifyLedgerTool struct{ *ledgerToolCtx }

func (t verifyLedgerTool) Name() string { return "verify_ledger" }

func (t verifyLedgerTool) Description() string {
	return `Verify Merkle integrity of a ledger. Input JSON: {"ledgerId":"..."} or plain ledger id. Uses bound ledger when omitted.`
}

func (t verifyLedgerTool) Call(ctx context.Context, input string) (string, error) {
	if t.svc == nil {
		return "ledger service unavailable", nil
	}
	ledgerID, err := t.resolveLedgerID(input)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	if _, err := t.svc.GetForUser(ctx, ledgerID, t.userID); err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	ok, err := t.svc.Verify(ctx, ledgerID)
	if err != nil {
		return fmt.Sprintf("error: %v", err), nil
	}
	out := map[string]any{"ledgerId": ledgerID, "valid": ok}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func (t *ledgerToolCtx) resolveLedgerID(input string) (string, error) {
	var args struct {
		LedgerID string `json:"ledgerId"`
	}
	if err := parseToolInput(input, &args); err == nil && strings.TrimSpace(args.LedgerID) != "" {
		return strings.TrimSpace(args.LedgerID), nil
	}
	plain := strings.TrimSpace(input)
	if plain != "" && !strings.HasPrefix(plain, "{") {
		return plain, nil
	}
	if t.defaultLedgerID != "" {
		return t.defaultLedgerID, nil
	}
	return "", fmt.Errorf("ledgerId required")
}

func parseToolInput(input string, dest any) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	return json.Unmarshal([]byte(input), dest)
}
