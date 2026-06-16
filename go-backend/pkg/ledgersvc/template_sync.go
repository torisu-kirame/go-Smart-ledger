package ledgersvc

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

// SyncEntryTemplateSchema updates entrySchema on all simple ledgers (and tables)
// that reference templateID for members who can edit the ledger (creator).
func (s *Service) SyncEntryTemplateSchema(ctx context.Context, userID, templateID string, schema domain.EntrySchema) (int, error) {
	if userID == "" || templateID == "" || templateID == "custom" {
		return 0, nil
	}
	schema.TemplateID = templateID
	if err := domain.ValidateSchema(schema); err != nil {
		return 0, err
	}
	list, err := s.ListForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, meta := range list {
		if meta == nil || domain.ResolvedBookkeepingMode(meta) != domain.BookkeepingSimple {
			continue
		}
		if meta.CreatorID != userID {
			continue
		}
		domain.NormalizeLedgerTables(meta)
		changed := false
		if meta.EntrySchema.TemplateID == templateID {
			meta.EntrySchema = schema
			changed = true
		}
		for i := range meta.Tables {
			if meta.Tables[i].EntrySchema.TemplateID == templateID {
				meta.Tables[i].EntrySchema = schema
				changed = true
			}
		}
		if !changed {
			continue
		}
		if err := s.putMeta(ctx, meta); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
