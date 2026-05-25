package friendstore

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

var (
	ErrFriendNotFound  = errors.New("friend not found")
	ErrCannotAddSelf   = errors.New("cannot add yourself")
	ErrAlreadyFriends  = errors.New("already friends")
	ErrFriendNotExists = errors.New("target user not found")
)

type Friend struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarUrl string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(userID string) ([]Friend, error) {
	uid, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
		SELECT u.id, u.username, COALESCE(NULLIF(u.nickname, ''), u.username),
		       COALESCE(NULLIF(u.avatar_url, ''), CONCAT('/api/v1/users/', u.id, '/avatar')),
		       f.created_at
		FROM friendships f
		INNER JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Friend
	for rows.Next() {
		var id uint64
		var username, nickname, avatarURL string
		var created time.Time
		if err := rows.Scan(&id, &username, &nickname, &avatarURL, &created); err != nil {
			return nil, err
		}
		out = append(out, Friend{
			ID:        strconv.FormatUint(id, 10),
			Username:  username,
			Nickname:  nickname,
			AvatarUrl: avatarURL,
			CreatedAt: created,
		})
	}
	return out, rows.Err()
}

func (s *Store) AreFriends(userID, friendID string) (bool, error) {
	uid, err := parseID(userID)
	if err != nil {
		return false, err
	}
	fid, err := parseID(friendID)
	if err != nil {
		return false, err
	}
	return s.areFriendsEither(uid, fid)
}

func (s *Store) Remove(userID, friendID string) error {
	uid, err := parseID(userID)
	if err != nil {
		return err
	}
	fid, err := parseID(friendID)
	if err != nil {
		return ErrFriendNotFound
	}
	res, err := s.db.Exec(
		`DELETE FROM friendships WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)`,
		uid, fid, fid, uid,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrFriendNotFound
	}
	return nil
}

func parseID(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
