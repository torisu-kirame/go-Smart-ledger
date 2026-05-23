package mapper

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
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
	fields := make([]domain.EntryFieldDef, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = domain.EntryFieldDef{
			Key:      f.Key,
			Label:    f.Label,
			Type:     domain.FieldType(f.Type),
			Required: f.Required,
		}
	}
	return domain.EntrySchema{TemplateID: s.TemplateId, Fields: fields}
}
