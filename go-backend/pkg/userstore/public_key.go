package userstore

import (
	"database/sql"
	"strconv"
)

// PublicKeyStore persists user encryption public keys (F19).
type PublicKeyStore interface {
	SetPublicKey(userID, pem string) error
	GetPublicKey(userID string) (string, error)
}

func (s *MySQLStore) SetPublicKey(userID, pem string) error {
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return ErrUserNotFound
	}
	_, err = s.db.Exec(`
		INSERT INTO user_public_keys (user_id, public_key_pem) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE public_key_pem = VALUES(public_key_pem)`,
		uid, pem)
	return err
}

func (s *MySQLStore) GetPublicKey(userID string) (string, error) {
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", ErrUserNotFound
	}
	var pem string
	err = s.db.QueryRow(`SELECT public_key_pem FROM user_public_keys WHERE user_id = ?`, uid).Scan(&pem)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return pem, err
}
