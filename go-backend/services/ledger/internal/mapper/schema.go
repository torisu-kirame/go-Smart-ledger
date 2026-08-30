package mapper

import (
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/types"
)

func EntrySchemaToResp(s domain.EntrySchema) types.EntrySchemaResp {
	fields := make([]types.EntryFieldDef, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = types.EntryFieldDef{
			Key:      f.Key,
			Label:    f.Label,
			Type:     string(f.Type),
			Required: f.Required,
		}
	}
	return types.EntrySchemaResp{
		TemplateId: s.TemplateID,
		Fields:     fields,
	}
}

func EntrySchemaFromReq(s types.EntrySchemaReq) domain.EntrySchema {
	fields := make([]domain.EntryFieldDef, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = domain.EntryFieldDef{
			Key:      f.Key,
			Label:    f.Label,
			Type:     domain.FieldType(f.Type),
			Required: f.Required,
		}
	}
	if s.TemplateId == domain.TemplateCustom {
		return domain.EntrySchema{TemplateID: domain.TemplateCustom, Fields: fields}
	}
	if len(s.Fields) == 0 && s.TemplateId == "" {
		return domain.EntrySchema{}
	}
	if len(s.Fields) == 0 {
		for _, t := range domain.BuiltinTemplates() {
			if t.TemplateID == s.TemplateId {
				return t
			}
		}
		return domain.DefaultEntrySchema()
	}
	return domain.EntrySchema{TemplateID: s.TemplateId, Fields: fields}
}

func TablesToResp(tables []domain.LedgerTable) []types.LedgerTableResp {
	if len(tables) == 0 {
		return nil
	}
	out := make([]types.LedgerTableResp, len(tables))
	for i, t := range tables {
		created := ""
		if !t.CreatedAt.IsZero() {
			created = t.CreatedAt.UTC().Format(time.RFC3339)
		}
		out[i] = types.LedgerTableResp{
			Id:          t.ID,
			Name:        t.Name,
			EntrySchema: EntrySchemaToResp(domain.ResolveEntrySchema(t.EntrySchema)),
			SortOrder:   t.SortOrder,
			RowOrder:    append([]uint64(nil), t.RowOrder...),
			CreatedAt:   created,
		}
	}
	return out
}
