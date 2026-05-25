package userstore

import (
	"strconv"
)

// DeleteAccount removes user after verifying username and password.
func (s *MySQLStore) DeleteAccount(id, username, password string) error {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ErrUserNotFound
	}
	user, err := s.Authenticate(username, password)
	if err != nil {
		return ErrInvalidCredentials
	}
	if user.ID != id {
		return ErrUsernameMismatch
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM friendships WHERE user_id = ? OR friend_id = ?`, uid, uid); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM friend_requests WHERE from_user_id = ? OR to_user_id = ?`, uid, uid); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM users WHERE id = ?`, uid)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return tx.Commit()
}
