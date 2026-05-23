package userstore

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
	"golang.org/x/crypto/bcrypt"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) Authenticate(username, password string) (*User, error) {
	var id uint64
	var hash string
	err := s.db.QueryRow(
		`SELECT id, password_hash FROM users WHERE username = ? LIMIT 1`,
		username,
	).Scan(&id, &hash)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return &User{ID: strconv.FormatUint(id, 10), Username: username}, nil
}

func (s *MySQLStore) Create(username, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	idStr, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(
		`INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)`,
		id, username, string(hash),
	)
	if err != nil {
		if isDuplicate(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return &User{ID: idStr, Username: username}, nil
}

func (s *MySQLStore) FindByID(id string) (*User, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, ErrUserNotFound
	}
	var username string
	err = s.db.QueryRow(
		`SELECT username FROM users WHERE id = ? LIMIT 1`,
		uid,
	).Scan(&username)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &User{ID: strconv.FormatUint(uid, 10), Username: username}, nil
}

// EnsureSeed creates the first admin user when the table is empty.
// Returns the new user's ID when created, or "" if users already exist.
func (s *MySQLStore) EnsureSeed(username, password string) (string, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return "", err
	}
	if n > 0 {
		return "", nil
	}
	u, err := s.Create(username, password)
	if err != nil {
		return "", err
	}
	return u.ID, nil
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate") || strings.Contains(msg, "1062")
}
