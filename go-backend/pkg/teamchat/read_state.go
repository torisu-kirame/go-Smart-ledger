package teamchat

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// LastMessagePreview is the latest chat line shown in team list.
type LastMessagePreview struct {
	ID             string    `json:"id,omitempty"`
	Type           string    `json:"type,omitempty"`
	Preview        string    `json:"preview"`
	SenderID       string    `json:"senderId,omitempty"`
	SenderNickname string    `json:"senderNickname,omitempty"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
}

func (s *Store) GetLastReadMessageID(teamID, userID string) (uint64, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return 0, err
	}
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return 0, err
	}
	var lastRead sql.NullInt64
	err = s.db.QueryRow(
		`SELECT last_read_message_id FROM team_read_state WHERE team_id = ? AND user_id = ?`,
		tid, uid,
	).Scan(&lastRead)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !lastRead.Valid {
		return 0, nil
	}
	return uint64(lastRead.Int64), nil
}

func (s *Store) LatestMessageID(teamID string) (uint64, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return 0, err
	}
	var id sql.NullInt64
	err = s.db.QueryRow(`SELECT MAX(id) FROM team_messages WHERE team_id = ?`, tid).Scan(&id)
	if err != nil {
		return 0, err
	}
	if !id.Valid {
		return 0, nil
	}
	return uint64(id.Int64), nil
}

// MarkTeamRead sets read cursor to the latest message in the team.
func (s *Store) MarkTeamRead(teamID, userID string) error {
	latest, err := s.LatestMessageID(teamID)
	if err != nil {
		return err
	}
	return s.setLastRead(teamID, userID, latest)
}

// MarkTeamReadUpTo sets read cursor at least to messageID.
func (s *Store) MarkTeamReadUpTo(teamID, userID, messageID string) error {
	mid, err := strconv.ParseUint(messageID, 10, 64)
	if err != nil {
		return ErrInvalidMessage
	}
	cur, _ := s.GetLastReadMessageID(teamID, userID)
	if mid > cur {
		return s.setLastRead(teamID, userID, mid)
	}
	return nil
}

func (s *Store) setLastRead(teamID, userID string, messageID uint64) error {
	tid, _ := strconv.ParseUint(teamID, 10, 64)
	uid, _ := strconv.ParseUint(userID, 10, 64)
	_, err := s.db.Exec(`
		INSERT INTO team_read_state (team_id, user_id, last_read_message_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE last_read_message_id = GREATEST(last_read_message_id, VALUES(last_read_message_id)),
			updated_at = CURRENT_TIMESTAMP`,
		tid, uid, messageID)
	return err
}

// MarkAllTeamsRead marks every team the user belongs to as read.
func (s *Store) MarkAllTeamsRead(userID string, teamIDs []string) error {
	for _, tid := range teamIDs {
		if err := s.MarkTeamRead(tid, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) UnreadCount(teamID, userID string) (int, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return 0, err
	}
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return 0, err
	}
	lastRead, err := s.GetLastReadMessageID(teamID, userID)
	if err != nil {
		return 0, err
	}
	var n int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM team_messages
		WHERE team_id = ? AND id > ? AND sender_id != ?`,
		tid, lastRead, uid,
	).Scan(&n)
	return n, err
}

func (s *Store) LastMessage(teamID string) (*LastMessagePreview, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, err
	}
	var msg Message
	var id, sid uint64
	var filePath string
	err = s.db.QueryRow(`
		SELECT m.id, m.sender_id, m.msg_type, COALESCE(m.body,''),
		       COALESCE(m.file_name,''), COALESCE(m.file_path,''), m.created_at,
		       COALESCE(NULLIF(u.nickname,''), u.username)
		FROM team_messages m
		INNER JOIN users u ON u.id = m.sender_id
		WHERE m.team_id = ?
		ORDER BY m.id DESC LIMIT 1`, tid,
	).Scan(&id, &sid, &msg.Type, &msg.Body, &msg.FileName, &filePath, &msg.CreatedAt, &msg.SenderNickname)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	msg.ID = strconv.FormatUint(id, 10)
	msg.SenderID = strconv.FormatUint(sid, 10)
	return &LastMessagePreview{
		ID:             msg.ID,
		Type:           msg.Type,
		Preview:        formatPreview(msg.Type, msg.Body, msg.FileName),
		SenderID:       msg.SenderID,
		SenderNickname: msg.SenderNickname,
		CreatedAt:      msg.CreatedAt,
	}, nil
}

func formatPreview(msgType, body, fileName string) string {
	switch msgType {
	case MsgTypeFile:
		if fileName != "" {
			return "[文件] " + fileName
		}
		return "[文件]"
	default:
		body = strings.TrimSpace(body)
		if body == "" {
			return ""
		}
		r := []rune(body)
		if len(r) > 60 {
			return string(r[:60]) + "…"
		}
		return body
	}
}
