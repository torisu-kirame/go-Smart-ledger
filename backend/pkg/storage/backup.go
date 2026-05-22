package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
)

var ErrNotFound = errors.New("backup not found")

// DiskBackup stores password-encrypted blobs on local disk.
type DiskBackup struct {
	mu   sync.Mutex
	root string
}

func NewDiskBackup(root string) (*DiskBackup, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, err
	}
	return &DiskBackup{root: root}, nil
}

func (d *DiskBackup) Put(_ context.Context, ledgerID, password string, plain []byte) (string, error) {
	cipher, err := encrypt(password, plain)
	if err != nil {
		return "", err
	}
	id, err := randomID()
	if err != nil {
		return "", err
	}
	ref := filepath.Join(ledgerID, id)
	path := filepath.Join(d.root, ref)
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, cipher, 0o600); err != nil {
		return "", err
	}
	return ref, nil
}

func (d *DiskBackup) Get(_ context.Context, ref, password string) ([]byte, error) {
	path := filepath.Join(d.root, ref)
	d.mu.Lock()
	data, err := os.ReadFile(path)
	d.mu.Unlock()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return decrypt(password, data)
}

func encrypt(password string, plain []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	out := gcm.Seal(nonce, nonce, plain, nil)
	return append(salt, out...), nil
}

func decrypt(password string, data []byte) ([]byte, error) {
	if len(data) < 16 {
		return nil, errors.New("invalid ciphertext")
	}
	salt := data[:16]
	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	body := data[16:]
	if len(body) < gcm.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}
	nonce := body[:gcm.NonceSize()]
	return gcm.Open(nil, nonce, body[gcm.NonceSize():], nil)
}

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 2, 64*1024, 4, 32)
}

func randomID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// EncodeB64 helper for API layer.
func EncodeB64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func DecodeB64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func ValidateRef(ref string) error {
	if ref == "" || ref == "." || ref == ".." {
		return fmt.Errorf("invalid ref")
	}
	return nil
}

// HashPasswordFingerprint for audit logs only.
func HashPasswordFingerprint(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:8])
}
