package teamstore

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

var (
	ErrTeamNotFound   = errors.New("team not found")
	ErrInvalidTeam    = errors.New("invalid team request")
	ErrNotFriend      = errors.New("member must be your friend")
	ErrLedgerNotMulti = errors.New("team requires a multi-person ledger")
	ErrNeedFriend     = errors.New("at least one friend required")
)

type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	LedgerID  string    `json:"ledgerId"`
	CreatorID string    `json:"creatorId"`
	CreatedAt time.Time `json:"createdAt"`
	Members   []Member  `json:"members"`
}

type Member struct {
	UserID   string `json:"userId"`
	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateWithID persists team; memberUserIDs must be friends of creatorID (validated by caller).
func (s *Store) CreateWithID(teamID, name, ledgerID, creatorID string, memberUserIDs []string) (*Team, error) {
	if name == "" || ledgerID == "" || creatorID == "" {
		return nil, ErrInvalidTeam
	}
	if len(memberUserIDs) < 1 {
		return nil, ErrNeedFriend
	}
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, ErrInvalidTeam
	}
	cid, err := strconv.ParseUint(creatorID, 10, 64)
	if err != nil {
		return nil, ErrInvalidTeam
	}
	seen := map[uint64]bool{cid: true}
	var mids []uint64
	for _, uid := range memberUserIDs {
		mid, err := strconv.ParseUint(uid, 10, 64)
		if err != nil || seen[mid] {
			return nil, ErrInvalidTeam
		}
		seen[mid] = true
		mids = append(mids, mid)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO teams (id, name, ledger_id, creator_id) VALUES (?, ?, ?, ?)`,
		tid, name, ledgerID, cid,
	); err != nil {
		return nil, err
	}
	for _, mid := range mids {
		if _, err := tx.Exec(`INSERT INTO team_members (team_id, user_id) VALUES (?, ?)`, tid, mid); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetByID(teamID)
}

func (s *Store) ListByUser(userID string) ([]Team, error) {
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, ErrInvalidTeam
	}
	rows, err := s.db.Query(`
		SELECT DISTINCT t.id, t.name, t.ledger_id, t.creator_id, t.created_at
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		WHERE t.creator_id = ? OR tm.user_id = ?
		ORDER BY t.created_at DESC`, uid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Team
	for rows.Next() {
		var t Team
		var id, creator uint64
		if err := rows.Scan(&id, &t.Name, &t.LedgerID, &creator, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.ID = strconv.FormatUint(id, 10)
		t.CreatorID = strconv.FormatUint(creator, 10)
		t.Members, _ = s.listMembers(t.ID)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetByID(teamID string) (*Team, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, ErrTeamNotFound
	}
	var t Team
	var id, creator uint64
	err = s.db.QueryRow(
		`SELECT id, name, ledger_id, creator_id, created_at FROM teams WHERE id = ?`, tid,
	).Scan(&id, &t.Name, &t.LedgerID, &creator, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}
	t.ID = strconv.FormatUint(id, 10)
	t.CreatorID = strconv.FormatUint(creator, 10)
	t.Members, _ = s.listMembers(t.ID)
	return &t, nil
}

func (s *Store) listMembers(teamID string) ([]Member, error) {
	tid, _ := strconv.ParseUint(teamID, 10, 64)
	rows, err := s.db.Query(`
		SELECT u.id, u.username, COALESCE(NULLIF(u.nickname, ''), u.username)
		FROM (
			SELECT creator_id AS user_id FROM teams WHERE id = ?
			UNION
			SELECT user_id FROM team_members WHERE team_id = ?
		) x
		INNER JOIN users u ON u.id = x.user_id
	`, tid, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		var id uint64
		if err := rows.Scan(&id, &m.Username, &m.Nickname); err != nil {
			return nil, err
		}
		m.UserID = strconv.FormatUint(id, 10)
		out = append(out, m)
	}
	return out, rows.Err()
}
