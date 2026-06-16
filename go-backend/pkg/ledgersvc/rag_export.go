package ledgersvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

// RAGChunk is one embeddable document unit for offline vector indexing.
type RAGChunk struct {
	ID        string         `json:"id"`
	LedgerID  string         `json:"ledgerId"`
	Seq       uint64         `json:"seq"`
	Type      string         `json:"type"`
	Text      string         `json:"text"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// RAGExport bundles ledger context for private local RAG (OpenClaw / LanceDB).
type RAGExport struct {
	LedgerID   string     `json:"ledgerId"`
	LedgerName string     `json:"ledgerName"`
	ExportedAt time.Time  `json:"exportedAt"`
	Chunks     []RAGChunk `json:"chunks"`
}

// ExportRAG builds text chunks for ledgers the user is allowed to see.
func (s *Service) ExportRAG(ctx context.Context, ledgerID, userID string) (*RAGExport, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	events, err := s.ListEvents(ctx, ledgerID, 1, meta.LatestSeq)
	if err != nil {
		return nil, err
	}
	out := &RAGExport{
		LedgerID:   ledgerID,
		LedgerName: meta.Name,
		ExportedAt: time.Now().UTC(),
		Chunks:     make([]RAGChunk, 0, len(events)+1),
	}
	out.Chunks = append(out.Chunks, RAGChunk{
		ID:       fmt.Sprintf("%s:meta", ledgerID),
		LedgerID: ledgerID,
		Type:     "ledger_meta",
		Text: fmt.Sprintf("账本「%s」(%s)，类型 %s，成员 %d 人，最新序号 %d，Merkle 根 %s，锚定状态 %s。",
			meta.Name, ledgerID, meta.Type, len(meta.Members), meta.LatestSeq, meta.LatestRoot, meta.AnchorStatus),
		Metadata: map[string]any{"latestSeq": meta.LatestSeq, "anchorStatus": meta.AnchorStatus},
		CreatedAt: meta.UpdatedAt,
	})
	schema := domain.ResolveEntrySchema(meta.EntrySchema)
	for _, ev := range events {
		ch := RAGChunk{
			ID:        fmt.Sprintf("%s:%d", ledgerID, ev.Seq),
			LedgerID:  ledgerID,
			Seq:       ev.Seq,
			Type:      ev.Type,
			CreatedAt: ev.CreatedAt,
		}
		ch.Text = eventToRAGText(ev, schema)
		out.Chunks = append(out.Chunks, ch)
	}
	return out, nil
}

func eventToRAGText(ev domain.EventRecord, schema domain.EntrySchema) string {
	switch ev.Type {
	case domain.EventEntryAdded:
		var entry domain.EntryPayload
		if json.Unmarshal(ev.Payload, &entry) == nil {
			data := entry.NormalizeData()
			parts := []string{fmt.Sprintf("记账事件 #%d", ev.Seq)}
			for _, f := range schema.Fields {
				if v := data[f.Key]; v != "" {
					parts = append(parts, fmt.Sprintf("%s: %s", f.Label, v))
				}
			}
			return strings.Join(parts, "；")
		}
	case domain.EventBatchSealed:
		return fmt.Sprintf("封账锚定 #%d", ev.Seq)
	case domain.EventExternalAnchored:
		return fmt.Sprintf("链外 EVM 锚定 #%d", ev.Seq)
	case domain.EventBackupAnchored:
		return fmt.Sprintf("备份已上链记录 #%d", ev.Seq)
	case domain.EventImportBatch:
		return fmt.Sprintf("批量导入 #%d", ev.Seq)
	}
	if len(ev.Payload) > 0 {
		return fmt.Sprintf("%s #%d: %s", ev.Type, ev.Seq, string(ev.Payload))
	}
	return fmt.Sprintf("%s #%d", ev.Type, ev.Seq)
}
