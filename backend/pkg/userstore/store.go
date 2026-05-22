package userstore

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type User struct {
	ID       string
	Username string
	hash     []byte
}

// MemoryStore is a simple in-memory user store for bootstrap.
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

type SeedUser struct {
	Username string
	Password string
}

func NewMemoryStore(seed []SeedUser) (*MemoryStore, error) {
	s := &MemoryStore{users: make(map[string]*User)}
	for i, u := range seed {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		id := u.Username
		if id == "" {
			id = "user"
		}
		s.users[u.Username] = &User{ID: id, Username: u.Username, hash: hash}
		_ = i
	}
	return s, nil
}

func (s *MemoryStore) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	u, ok := s.users[username]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword(u.hash, []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}
