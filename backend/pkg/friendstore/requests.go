package friendstore

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

const (
	StatusPending  = "pending"
	StatusRejected = "rejected"
)

var (
	ErrRequestPending    = errors.New("friend request already pending")
	ErrRequestNotFound   = errors.New("friend request not found")
	ErrIncomingPending = errors.New("incoming friend request pending")
)

// FriendRequest is a pending or historical friend invite.
type FriendRequest struct {
	FromUserId string    `json:"fromUserId"`
	ToUserId   string    `json:"toUserId"`
	Status     string    `json:"status"`
	Username   string    `json:"username,omitempty"`
	Nickname   string    `json:"nickname,omitempty"`
	AvatarUrl  string    `json:"avatarUrl,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// SendRequest creates a pending friend request (replaces instant Add).
func (s *Store) SendRequest(fromUserID, toUserID string) error {
	from, err := parseID(fromUserID)
	if err != nil {
		return err
	}
	to, err := parseID(toUserID)
	if err != nil {
		return ErrFriendNotExists
	}
	if from == to {
		return ErrCannotAddSelf
	}
	if err := s.userExists(to); err != nil {
		return err
	}
	if ok, err := s.areFriendsEither(from, to); err != nil {
		return err
	} else if ok {
		return ErrAlreadyFriends
	}
	var inc int
	err = s.db.QueryRow(
		`SELECT COUNT(*) FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? AND status = ?`,
		to, from, StatusPending,
	).Scan(&inc)
	if err != nil {
		return err
	}
	if inc > 0 {
		return ErrIncomingPending
	}
	var out int
	err = s.db.QueryRow(
		`SELECT COUNT(*) FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? AND status = ?`,
		from, to, StatusPending,
	).Scan(&out)
	if err != nil {
		return err
	}
	if out > 0 {
		return ErrRequestPending
	}
	_, err = s.db.Exec(
		`INSERT INTO friend_requests (from_user_id, to_user_id, status) VALUES (?, ?, ?)`,
		from, to, StatusPending,
	)
	return err
}

// AcceptRequest accepts an incoming request and creates mutual friendships.
func (s *Store) AcceptRequest(accepterID, fromUserID string) error {
	accepter, err := parseID(accepterID)
	if err != nil {
		return err
	}
	from, err := parseID(fromUserID)
	if err != nil {
		return ErrRequestNotFound
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow(
		`SELECT status FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? FOR UPDATE`,
		from, accepter,
	).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrRequestNotFound
	}
	if err != nil {
		return err
	}
	if status != StatusPending {
		return ErrRequestNotFound
	}
	if _, err := tx.Exec(`DELETE FROM friend_requests WHERE from_user_id = ? AND to_user_id = ?`, from, accepter); err != nil {
		return err
	}
	if err := insertFriendshipPair(tx, from, accepter); err != nil {
		return err
	}
	return tx.Commit()
}

// RejectRequest declines an incoming request.
func (s *Store) RejectRequest(accepterID, fromUserID string) error {
	accepter, err := parseID(accepterID)
	if err != nil {
		return err
	}
	from, err := parseID(fromUserID)
	if err != nil {
		return ErrRequestNotFound
	}
	res, err := s.db.Exec(
		`DELETE FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? AND status = ?`,
		from, accepter, StatusPending,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// CancelRequest withdraws an outgoing pending request.
func (s *Store) CancelRequest(fromUserID, toUserID string) error {
	from, err := parseID(fromUserID)
	if err != nil {
		return err
	}
	to, err := parseID(toUserID)
	if err != nil {
		return ErrRequestNotFound
	}
	res, err := s.db.Exec(
		`DELETE FROM friend_requests WHERE from_user_id = ? AND to_user_id = ? AND status = ?`,
		from, to, StatusPending,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRequestNotFound
	}
	return nil
}

func (s *Store) ListIncomingRequests(userID string) ([]FriendRequest, error) {
	uid, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	return s.listRequests(`fr.to_user_id = ? AND fr.status = ?`, uid, StatusPending, true)
}

func (s *Store) ListOutgoingRequests(userID string) ([]FriendRequest, error) {
	uid, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	return s.listRequests(`fr.from_user_id = ? AND fr.status = ?`, uid, StatusPending, false)
}

func (s *Store) listRequests(whereClause string, uid interface{}, status string, incoming bool) ([]FriendRequest, error) {
	// incoming: join on from_user_id to show requester profile
	otherCol := "fr.from_user_id"
	if !incoming {
		otherCol = "fr.to_user_id"
	}
	rows, err := s.db.Query(`
		SELECT fr.from_user_id, fr.to_user_id, fr.status, fr.created_at,
		       u.id, u.username, COALESCE(NULLIF(u.nickname, ''), u.username),
		       COALESCE(NULLIF(u.avatar_url, ''), CONCAT('/api/v1/users/', u.id, '/avatar'))
		FROM friend_requests fr
		INNER JOIN users u ON u.id = `+otherCol+`
		WHERE `+whereClause+`
		ORDER BY fr.created_at DESC`, uid, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FriendRequest
	for rows.Next() {
		var fromID, toID uint64
		var st string
		var created time.Time
		var oid uint64
		var username, nickname, avatar string
		if err := rows.Scan(&fromID, &toID, &st, &created, &oid, &username, &nickname, &avatar); err != nil {
			return nil, err
		}
		fr := FriendRequest{
			FromUserId: strconv.FormatUint(fromID, 10),
			ToUserId:   strconv.FormatUint(toID, 10),
			Status:     st,
			Username:   username,
			Nickname:   nickname,
			AvatarUrl:  avatar,
			CreatedAt:  created,
		}
		out = append(out, fr)
	}
	return out, rows.Err()
}

func (s *Store) userExists(uid uint64) error {
	var dummy int
	err := s.db.QueryRow(`SELECT 1 FROM users WHERE id = ? LIMIT 1`, uid).Scan(&dummy)
	if err == sql.ErrNoRows {
		return ErrFriendNotExists
	}
	return err
}

func (s *Store) areFriendsEither(a, b uint64) (bool, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM friendships WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)`,
		a, b, b, a,
	).Scan(&n)
	return n > 0, err
}

func insertFriendshipPair(tx *sql.Tx, a, b uint64) error {
	for _, pair := range [][2]uint64{{a, b}, {b, a}} {
		var n int
		if err := tx.QueryRow(
			`SELECT COUNT(*) FROM friendships WHERE user_id = ? AND friend_id = ?`,
			pair[0], pair[1],
		).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO friendships (user_id, friend_id) VALUES (?, ?)`,
			pair[0], pair[1],
		); err != nil {
			return err
		}
	}
	return nil
}
