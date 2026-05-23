package userstore

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameMismatch   = errors.New("username does not match current account")
)

type User struct {
	ID       string
	Username string
	hash     []byte
}

// Store abstracts user persistence.
type Store interface {
	Authenticate(username, password string) (*User, error)
	Create(username, password string) (*User, error)
	FindByID(id string) (*User, error)
}

// AccountStore supports account lifecycle (delete).
type AccountStore interface {
	DeleteAccount(id, username, password string) error
}
