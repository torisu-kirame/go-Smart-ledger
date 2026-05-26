package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FieldType string

const (
	FieldText   FieldType = "text"
	FieldNumber FieldType = "number"
	FieldDate   FieldType = "date"
	FieldUser   FieldType = "user" // 账本成员用户 ID，链上 signer
)

// EntryFieldDef describes one column in ledger entry data.
type EntryFieldDef struct {
	Key      string    `json:"key"`
	Label    string    `json:"label"`
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
}

// EntrySchema is stored on ledger meta; drives forms and Excel import.
type EntrySchema struct {
	TemplateID string          `json:"templateId,omitempty"`
	Fields     []EntryFieldDef `json:"fields"`
}

const (
	TemplateDefault = "default"
	TemplateClassic = "classic"
)

var (
	ErrInvalidSchema   = errors.New("invalid entry schema")
	ErrEntryValidation = errors.New("entry validation failed")
)

// DefaultEntrySchema 默认模板：记账人、收账人、金额、日期、备注。
func DefaultEntrySchema() EntrySchema {
	return EntrySchema{
		TemplateID: TemplateDefault,
		Fields: []EntryFieldDef{
			{Key: "bookkeeper", Label: "记账人", Type: FieldUser, Required: true},
			{Key: "payee", Label: "收账人", Type: FieldText, Required: false},
			{Key: "amount", Label: "金额", Type: FieldNumber, Required: true},
			{Key: "date", Label: "日期", Type: FieldDate, Required: true},
			{Key: "note", Label: "备注", Type: FieldText, Required: false},
		},
	}
}

// ClassicEntrySchema 旧版导入列（兼容已有账本）。
func ClassicEntrySchema() EntrySchema {
	return EntrySchema{
		TemplateID: TemplateClassic,
		Fields: []EntryFieldDef{
			{Key: "date", Label: "日期", Type: FieldDate, Required: true},
			{Key: "type", Label: "类型", Type: FieldText, Required: true},
			{Key: "amount", Label: "金额", Type: FieldNumber, Required: true},
			{Key: "category", Label: "分类", Type: FieldText, Required: false},
			{Key: "note", Label: "备注", Type: FieldText, Required: false},
			{Key: "counterparty", Label: "对方", Type: FieldText, Required: false},
		},
	}
}

// BuiltinTemplates returns preset schemas for API / UI (excludes legacy classic).
func BuiltinTemplates() []EntrySchema {
	return []EntrySchema{DefaultEntrySchema()}
}

// ResolveEntrySchema returns ledger schema; empty meta uses classic columns (legacy ledgers).
func ResolveEntrySchema(s EntrySchema) EntrySchema {
	if s.TemplateID == TemplateProfessional {
		return s
	}
	if len(s.Fields) == 0 {
		return ClassicEntrySchema()
	}
	return s
}

// ValidateSchema checks field definitions.
func ValidateSchema(s EntrySchema) error {
	if len(s.Fields) == 0 {
		return ErrInvalidSchema
	}
	seen := map[string]bool{}
	for _, f := range s.Fields {
		k := strings.TrimSpace(f.Key)
		if k == "" || seen[k] {
			return ErrInvalidSchema
		}
		seen[k] = true
		if strings.TrimSpace(f.Label) == "" {
			return fmt.Errorf("%w: empty label for %s", ErrInvalidSchema, k)
		}
		switch f.Type {
		case FieldText, FieldNumber, FieldDate, FieldUser:
		default:
			return fmt.Errorf("%w: unknown type %s", ErrInvalidSchema, f.Type)
		}
	}
	return nil
}

// ValidateEntryData validates map against schema.
func ValidateEntryData(schema EntrySchema, data map[string]string) error {
	schema = ResolveEntrySchema(schema)
	for _, f := range schema.Fields {
		v := strings.TrimSpace(data[f.Key])
		if f.Required && v == "" {
			return fmt.Errorf("%w: %s 不能为空", ErrEntryValidation, f.Label)
		}
		if v == "" {
			continue
		}
		switch f.Type {
		case FieldNumber:
			if _, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", ""), 64); err != nil {
				return fmt.Errorf("%w: %s 格式无效", ErrEntryValidation, f.Label)
			}
		case FieldDate:
			if len(v) < 8 {
				return fmt.Errorf("%w: %s 格式无效", ErrEntryValidation, f.Label)
			}
		}
		if f.Key == "type" && schema.TemplateID == TemplateClassic {
			t := strings.ToLower(v)
			if t != "income" && t != "expense" && t != "收入" && t != "支出" {
				return fmt.Errorf("%w: 类型须为 收入/支出", ErrEntryValidation)
			}
		}
	}
	return nil
}

// SignerFromEntry picks chain signer: FieldUser key (e.g. bookkeeper) or explicit signerId.
func SignerFromEntry(schema EntrySchema, data map[string]string, explicitSigner string) (string, error) {
	if explicitSigner != "" {
		return explicitSigner, nil
	}
	schema = ResolveEntrySchema(schema)
	for _, f := range schema.Fields {
		if f.Type == FieldUser {
			if v := strings.TrimSpace(data[f.Key]); v != "" {
				return v, nil
			}
		}
	}
	return "", fmt.Errorf("%w: 缺少记账人", ErrEntryValidation)
}
