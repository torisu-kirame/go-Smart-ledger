package userstore

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
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
