package userstore

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// MemoryStore is a simple in-memory user store for bootstrap without MySQL.
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]*User
	byID  map[string]*User
}

type SeedUser struct {
	Username string
	Password string
}

func NewMemoryStore(seed []SeedUser) (*MemoryStore, error) {
	s := &MemoryStore{
		users: make(map[string]*User),
		byID:  make(map[string]*User),
	}
	for _, u := range seed {
		if _, err := s.Create(u.Username, u.Password); err != nil {
			return nil, err
		}
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
	return &User{ID: u.ID, Username: u.Username}, nil
}

func (s *MemoryStore) Create(username, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[username]; ok {
		return nil, ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id := username
	u := &User{ID: id, Username: username, hash: hash}
	s.users[username] = u
	s.byID[id] = u
	return &User{ID: u.ID, Username: u.Username}, nil
}

func (s *MemoryStore) FindByID(id string) (*User, error) {
	s.mu.RLock()
	u, ok := s.byID[id]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrUserNotFound
	}
	return &User{ID: u.ID, Username: u.Username}, nil
}
