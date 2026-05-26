package entrytemplatestore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

var (
	ErrNotFound    = errors.New("template not found")
	ErrForbidden   = errors.New("forbidden")
	ErrInvalid     = errors.New("invalid template")
	ErrBuiltinRead = errors.New("builtin template cannot be modified")
)

// Template is a user-saved or built-in entry schema.
type Template struct {
	TemplateID string                `json:"templateId"`
	Name       string                `json:"name"`
	Builtin    bool                  `json:"builtin"`
	OwnerID    string                `json:"ownerId,omitempty"`
	Fields     []domain.EntryFieldDef `json:"fields"`
	CreatedAt  time.Time             `json:"createdAt,omitempty"`
	UpdatedAt  time.Time             `json:"updatedAt,omitempty"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListForUser(userID string) ([]Template, error) {
	out := builtinTemplates()
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return out, nil
	}
	rows, err := s.db.Query(
		`SELECT id, name, fields_json, created_at, updated_at FROM entry_templates WHERE owner_id = ? ORDER BY updated_at DESC`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t Template
		var id uint64
		var raw []byte
		if err := rows.Scan(&id, &t.Name, &raw, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.TemplateID = strconv.FormatUint(id, 10)
		t.OwnerID = userID
		if err := json.Unmarshal(raw, &t.Fields); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetByID(templateID, userID string) (*Template, error) {
	if templateID == domain.TemplateClassic {
		// Legacy built-in removed from listing; still resolvable for old ledgers.
		sch := domain.ClassicEntrySchema()
		now := time.Now().UTC()
		return &Template{
			TemplateID: sch.TemplateID,
			Name:       "经典记账（已下线）",
			Builtin:    true,
			Fields:     sch.Fields,
			CreatedAt:  now,
		}, nil
	}
	if isBuiltinID(templateID) {
		for _, t := range builtinTemplates() {
			if t.TemplateID == templateID {
				return &t, nil
			}
		}
		return nil, ErrNotFound
	}
	queryID, err := strconv.ParseUint(templateID, 10, 64)
	if err != nil {
		return nil, ErrNotFound
	}
	var t Template
	var raw []byte
	var id, owner uint64
	err = s.db.QueryRow(
		`SELECT id, owner_id, name, fields_json, created_at, updated_at FROM entry_templates WHERE id = ?`,
		queryID,
	).Scan(&id, &owner, &t.Name, &raw, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.TemplateID = strconv.FormatUint(id, 10)
	t.OwnerID = strconv.FormatUint(owner, 10)
	if userID != "" && t.OwnerID != userID {
		return nil, ErrForbidden
	}
	if err := json.Unmarshal(raw, &t.Fields); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) CreateWithID(id, ownerID, name string, fields []domain.EntryFieldDef) (*Template, error) {
	if name == "" || len(fields) == 0 {
		return nil, ErrInvalid
	}
	if err := domain.ValidateSchema(domain.EntrySchema{TemplateID: id, Fields: fields}); err != nil {
		return nil, err
	}
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, ErrInvalid
	}
	oid, err := strconv.ParseUint(ownerID, 10, 64)
	if err != nil {
		return nil, ErrInvalid
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(
		`INSERT INTO entry_templates (id, owner_id, name, fields_json) VALUES (?, ?, ?, ?)`,
		tid, oid, name, raw,
	)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id, ownerID)
}

func (s *Store) Update(templateID, ownerID, name string, fields []domain.EntryFieldDef) (*Template, error) {
	if isBuiltinID(templateID) {
		return nil, ErrBuiltinRead
	}
	if name == "" || len(fields) == 0 {
		return nil, ErrInvalid
	}
	if err := domain.ValidateSchema(domain.EntrySchema{TemplateID: templateID, Fields: fields}); err != nil {
		return nil, err
	}
	tid, err := strconv.ParseUint(templateID, 10, 64)
	if err != nil {
		return nil, ErrNotFound
	}
	oid, err := strconv.ParseUint(ownerID, 10, 64)
	if err != nil {
		return nil, ErrInvalid
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`UPDATE entry_templates SET name = ?, fields_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND owner_id = ?`,
		name, raw, tid, oid,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.GetByID(templateID, ownerID)
}

func (s *Store) Delete(templateID, ownerID string) error {
	if isBuiltinID(templateID) {
		return ErrBuiltinRead
	}
	tid, err := strconv.ParseUint(templateID, 10, 64)
	if err != nil {
		return ErrNotFound
	}
	oid, err := strconv.ParseUint(ownerID, 10, 64)
	if err != nil {
		return ErrInvalid
	}
	res, err := s.db.Exec(`DELETE FROM entry_templates WHERE id = ? AND owner_id = ?`, tid, oid)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ToEntrySchema(t *Template) domain.EntrySchema {
	return domain.EntrySchema{TemplateID: t.TemplateID, Fields: t.Fields}
}

func builtinTemplates() []Template {
	now := time.Now().UTC()
	out := make([]Template, 0, len(domain.BuiltinTemplates()))
	for _, s := range domain.BuiltinTemplates() {
		name := s.TemplateID
		if s.TemplateID == domain.TemplateDefault {
			name = "标准记账"
		}
		out = append(out, Template{
			TemplateID: s.TemplateID,
			Name:       name,
			Builtin:    true,
			Fields:     s.Fields,
			CreatedAt:  now,
		})
	}
	return out
}

func isBuiltinID(id string) bool {
	return id == domain.TemplateDefault
}
