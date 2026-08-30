package domain

import (
	"fmt"
	"strings"
)

// MergeEntrySchema appends fields from extra that are not already present in base
// (matched by field key or normalized label). Returns merged schema and a map from
// extra field key → final schema key (for remapping import row cells).
func MergeEntrySchema(base, extra EntrySchema) (EntrySchema, map[string]string) {
	base = ResolveEntrySchema(base)
	extra = ResolveEntrySchema(extra)
	out := EntrySchema{
		TemplateID: base.TemplateID,
		Fields:     append([]EntryFieldDef{}, base.Fields...),
	}
	if out.TemplateID == "" {
		out.TemplateID = extra.TemplateID
	}
	if out.TemplateID == "" {
		out.TemplateID = TemplateCustom
	}

	usedKey := map[string]bool{}
	labelToKey := map[string]string{}
	for _, f := range out.Fields {
		usedKey[f.Key] = true
		labelToKey[NormFieldLabel(f.Label)] = f.Key
		labelToKey[NormFieldLabel(f.Key)] = f.Key
	}

	remap := map[string]string{}
	for _, f := range extra.Fields {
		nk := strings.TrimSpace(f.Key)
		if nk == "" {
			continue
		}
		nl := NormFieldLabel(f.Label)
		if nl == "" {
			nl = NormFieldLabel(nk)
		}
		if existing, ok := labelToKey[nl]; ok {
			remap[nk] = existing
			continue
		}
		if usedKey[nk] {
			// key collision with different label — allocate unique key
			baseKey := nk
			for i := 2; usedKey[nk]; i++ {
				nk = fmt.Sprintf("%s_%d", baseKey, i)
			}
		}
		ff := f
		ff.Key = nk
		if strings.TrimSpace(ff.Label) == "" {
			ff.Label = nk
		}
		out.Fields = append(out.Fields, ff)
		usedKey[nk] = true
		labelToKey[NormFieldLabel(ff.Label)] = nk
		remap[f.Key] = nk
	}
	return out, remap
}

// NormFieldLabel normalizes labels for schema matching (spaces / unit parentheses).
func NormFieldLabel(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	if i := strings.IndexAny(s, "(（"); i > 0 {
		s = s[:i]
	}
	return strings.ToLower(s)
}

// RelaxRequiredForAbsentColumns clears Required on fields whose keys are not in
// presentKeys. Used when appending a file that only covers a subset of columns
// onto an existing sheet (e.g. target has required 日期/记账人 but file does not).
func RelaxRequiredForAbsentColumns(schema EntrySchema, presentKeys map[string]bool) EntrySchema {
	schema = ResolveEntrySchema(schema)
	if len(presentKeys) == 0 {
		return schema
	}
	out := schema
	fields := make([]EntryFieldDef, len(schema.Fields))
	for i, f := range schema.Fields {
		fields[i] = f
		if !presentKeys[f.Key] {
			fields[i].Required = false
		}
	}
	out.Fields = fields
	return out
}
