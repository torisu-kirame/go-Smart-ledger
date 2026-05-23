package userstore

import (
	"database/sql"
	"strconv"
	"time"
)

func (s *MySQLStore) GetProfile(id string) (*Profile, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, ErrUserNotFound
	}
	var username, nickname, avatarURL string
	var createdAt time.Time
	err = s.db.QueryRow(
		`SELECT username, COALESCE(nickname, ''), COALESCE(avatar_url, ''), created_at
		 FROM users WHERE id = ? LIMIT 1`,
		uid,
	).Scan(&username, &nickname, &avatarURL, &createdAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if avatarURL == "" {
		avatarURL = AvatarPath(id)
	}
	return &Profile{
		ID:        id,
		Username:  username,
		Nickname:  displayNickname(username, nickname),
		AvatarURL: avatarURL,
		CreatedAt: createdAt,
	}, nil
}

func (s *MySQLStore) UpdateNickname(id, nickname string) (*Profile, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, ErrUserNotFound
	}
	res, err := s.db.Exec(`UPDATE users SET nickname = ? WHERE id = ?`, nickname, uid)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrUserNotFound
	}
	return s.GetProfile(id)
}

func (s *MySQLStore) SetAvatarURL(id, avatarURL string) error {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ErrUserNotFound
	}
	res, err := s.db.Exec(`UPDATE users SET avatar_url = ? WHERE id = ?`, avatarURL, uid)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}
