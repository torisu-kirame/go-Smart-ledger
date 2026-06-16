package ledgersvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/auditexport"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

// BuildAuditBundle collects ledger data for F48 compliance export.
func (s *Service) BuildAuditBundle(ctx context.Context, ledgerID, userID string) (*auditexport.Bundle, error) {
	meta, err := s.GetForUser(ctx, ledgerID, userID)
	if err != nil {
		return nil, err
	}
	valid, _ := s.Verify(ctx, ledgerID)
	events, err := s.ListEvents(ctx, ledgerID, 1, meta.LatestSeq)
	if err != nil {
		return nil, err
	}
	domain.NormalizeLedgerTables(meta)
	bundle := &auditexport.Bundle{
		LedgerID:          ledgerID,
		LedgerName:        meta.Name,
		BookkeepingMode:   meta.BookkeepingMode,
		MultiTableEnabled: meta.MultiTableEnabled,
		LatestSeq:         meta.LatestSeq,
		LatestRoot:        meta.LatestRoot,
		AnchorStatus:      meta.AnchorStatus,
		IntegrityValid:    valid,
		ExportedAt:        time.Now().UTC(),
		ExportedBy:        userID,
	}
	if meta.ExternalAnchor != nil {
		bundle.ExternalAnchorTx = meta.ExternalAnchor.TxHash
	}

	tableMap := make(map[string]*auditexport.TableRows)
	for _, t := range meta.Tables {
		schema, err := domain.SchemaForTable(meta, t.ID)
		if err != nil {
			continue
		}
		tableMap[t.ID] = &auditexport.TableRows{
			TableID:   t.ID,
			TableName: t.Name,
			Schema:    schema,
			Rows:      nil,
		}
	}
	if len(tableMap) == 0 {
		schema := domain.ResolveEntrySchema(meta.EntrySchema)
		tableMap[domain.DefaultTableID] = &auditexport.TableRows{
			TableID:   domain.DefaultTableID,
			TableName: "默认",
			Schema:    schema,
		}
	}

	for _, ev := range events {
		if ev.Type != domain.EventEntryAdded {
			continue
		}
		var entry domain.EntryPayload
		if json.Unmarshal(ev.Payload, &entry) != nil {
			continue
		}
		tid := domain.EntryTableIDFromPayload(ev.Payload)
		if entry.TableID != "" {
			tid = domain.ResolveTableID(meta, entry.TableID)
		}
		tbl, ok := tableMap[tid]
		if !ok {
			schema, _ := domain.SchemaForTable(meta, tid)
			name := tid
			if t := domain.TableByID(meta, tid); t != nil {
				name = t.Name
			}
			tbl = &auditexport.TableRows{TableID: tid, TableName: name, Schema: schema}
			tableMap[tid] = tbl
		}
		tbl.Rows = append(tbl.Rows, auditexport.EntryRow{
			Seq:      ev.Seq,
			SignerID: ev.SignerID,
			Data:     entry.NormalizeData(),
			Hash:     ev.Hash,
			At:       ev.CreatedAt,
		})
	}
	for _, t := range meta.Tables {
		if tr, ok := tableMap[t.ID]; ok {
			bundle.Tables = append(bundle.Tables, *tr)
		}
	}
	if len(bundle.Tables) == 0 {
		for _, tr := range tableMap {
			bundle.Tables = append(bundle.Tables, *tr)
		}
	}

	atts, err := s.ListEntryAttachments(ctx, ledgerID, userID, "", 0)
	if err == nil {
		bundle.Attachments = atts
	}

	if domain.IsProfessionalBookkeeping(meta) {
		if chart, err := s.GetChart(ctx, ledgerID, userID); err == nil {
			bundle.Chart = chart
		}
		if journals, err := s.ListJournals(ctx, ledgerID, userID); err == nil {
			bundle.Journals = journals
		}
		if periods, err := s.listPeriodStates(ctx, ledgerID); err == nil {
			bundle.Periods = periods
		}
	}
	return bundle, nil
}

func (s *Service) listPeriodStates(ctx context.Context, ledgerID string) ([]accounting.PeriodState, error) {
	rows, err := s.chain.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		domain.LedgerPeriodPrefix(ledgerID)+"%")
	if err != nil {
		return nil, err
	}
	var out []accounting.PeriodState
	for _, row := range rows {
		var ps accounting.PeriodState
		if unmarshalStateValue(row.Value, &ps) == nil {
			out = append(out, ps)
		}
	}
	return out, nil
}

// ExportAudit returns rendered audit package bytes.
func (s *Service) ExportAudit(ctx context.Context, ledgerID, userID, format string) ([]byte, string, string, error) {
	bundle, err := s.BuildAuditBundle(ctx, ledgerID, userID)
	if err != nil {
		return nil, "", "", err
	}
	stamp := bundle.ExportedAt.Format("20060102-150405")
	switch format {
	case "pdf":
		data, err := auditexport.BuildPDF(bundle)
		return data, "application/pdf", "audit-" + stamp + "-summary.pdf", err
	case "zip":
		data, err := auditexport.BuildZip(bundle)
		return data, "application/zip", "audit-" + stamp + ".zip", err
	default:
		data, err := auditexport.BuildExcel(bundle)
		return data, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "audit-" + stamp + ".xlsx", err
	}
}
