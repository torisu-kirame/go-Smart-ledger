package teamchat

import (
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
)

const (
	MsgTypeText = "text"
	MsgTypeFile = "file"

	maxBodyLen  = 4000
	maxFileSize = 15 << 20 // 15 MiB
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrInvalidMessage  = errors.New("invalid message")
	ErrFileTooLarge    = errors.New("file too large")
	ErrFileType        = errors.New("file type not allowed")
)

// Message is a team chat message (F37).
type Message struct {
	ID             string    `json:"id"`
	TeamID         string    `json:"teamId"`
	SenderID       string    `json:"senderId"`
	SenderUsername string    `json:"senderUsername,omitempty"`
	SenderNickname string    `json:"senderNickname,omitempty"`
	Type           string    `json:"type"`
	Body           string    `json:"body,omitempty"`
	FileName       string    `json:"fileName,omitempty"`
	FileURL        string    `json:"fileUrl,omitempty"`
	FileSize       int64     `json:"fileSize,omitempty"`
	ContentType    string    `json:"contentType,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Store struct {
	db      *sql.DB
	fileDir string
}

func New(db *sql.DB, fileDir string) *Store {
	return &Store{db: db, fileDir: fileDir}
}

func FileAPIPath(teamID, messageID string) string {
	return "/api/v1/teams/" + teamID + "/chat/files/" + messageID
}

func (s *Store) List(teamID string, sinceID uint64, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, ErrInvalidMessage
	}
	q := `
		SELECT m.id, m.sender_id, m.msg_type, COALESCE(m.body,''),
		       COALESCE(m.file_name,''), COALESCE(m.file_path,''), m.file_size,
		       COALESCE(m.content_type,''), m.created_at,
		       u.username, COALESCE(NULLIF(u.nickname,''), u.username)
		FROM team_messages m
		INNER JOIN users u ON u.id = m.sender_id
		WHERE m.team_id = ?`
	args := []any{tid}
	if sinceID > 0 {
		q += ` AND m.id > ? ORDER BY m.id ASC LIMIT ?`
		args = append(args, sinceID, limit)
	} else {
		q += ` ORDER BY m.id DESC LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		msg, err := scanRow(rows, teamID)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	if sinceID == 0 && len(out) > 1 {
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return out, rows.Err()
}

func scanRow(rows *sql.Rows, teamID string) (Message, error) {
	var msg Message
	var id, sid uint64
	var filePath string
	if err := rows.Scan(
		&id, &sid, &msg.Type, &msg.Body,
		&msg.FileName, &filePath, &msg.FileSize, &msg.ContentType, &msg.CreatedAt,
		&msg.SenderUsername, &msg.SenderNickname,
	); err != nil {
		return msg, err
	}
	msg.ID = strconv.FormatUint(id, 10)
	msg.TeamID = teamID
	msg.SenderID = strconv.FormatUint(sid, 10)
	if msg.Type == MsgTypeFile && filePath != "" {
		msg.FileURL = FileAPIPath(teamID, msg.ID)
	}
	return msg, nil
}

func (s *Store) PostText(teamID, senderID, body string) (*Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > maxBodyLen {
		return nil, ErrInvalidMessage
	}
	return s.insertMessage(teamID, senderID, MsgTypeText, body, "", "", 0, "")
}

func (s *Store) PostFile(teamID, senderID, fileName, contentType string, r io.Reader) (*Message, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, ErrInvalidMessage
	}
	if !allowedContentType(contentType, fileName) {
		return nil, ErrFileType
	}
	idStr, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	relPath := filepath.Join(teamID, idStr+"_"+sanitizeFileName(fileName))
	absPath := filepath.Join(s.fileDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(absPath)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(f, io.LimitReader(r, maxFileSize+1))
	_ = f.Close()
	if err != nil {
		_ = os.Remove(absPath)
		return nil, err
	}
	if n > maxFileSize {
		_ = os.Remove(absPath)
		return nil, ErrFileTooLarge
	}
	if n == 0 {
		_ = os.Remove(absPath)
		return nil, ErrInvalidMessage
	}
	return s.insertMessageWithID(idStr, teamID, senderID, MsgTypeFile, "", fileName, relPath, n, contentType)
}

func (s *Store) insertMessage(teamID, senderID, msgType, body, fileName, filePath string, fileSize int64, contentType string) (*Message, error) {
	idStr, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	return s.insertMessageWithID(idStr, teamID, senderID, msgType, body, fileName, filePath, fileSize, contentType)
}

func (s *Store) insertMessageWithID(idStr, teamID, senderID, msgType, body, fileName, filePath string, fileSize int64, contentType string) (*Message, error) {
	id, _ := strconv.ParseUint(idStr, 10, 64)
	tid, _ := strconv.ParseUint(teamID, 10, 64)
	sid, _ := strconv.ParseUint(senderID, 10, 64)
	if tid == 0 || sid == 0 {
		return nil, ErrInvalidMessage
	}
	_, err := s.db.Exec(`
		INSERT INTO team_messages (id, team_id, sender_id, msg_type, body, file_name, file_path, file_size, content_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, tid, sid, msgType, body, fileName, filePath, fileSize, contentType,
	)
	if err != nil {
		if msgType == MsgTypeFile && filePath != "" {
			_ = os.Remove(filepath.Join(s.fileDir, filePath))
		}
		return nil, err
	}
	msgs, err := s.List(teamID, id-1, 1)
	if err == nil && len(msgs) > 0 {
		return &msgs[0], nil
	}
	m := &Message{
		ID: idStr, TeamID: teamID, SenderID: senderID, Type: msgType,
		Body: body, FileName: fileName, FileSize: fileSize, ContentType: contentType,
		CreatedAt: time.Now().UTC(),
	}
	if msgType == MsgTypeFile && filePath != "" {
		m.FileURL = FileAPIPath(teamID, idStr)
	}
	return m, nil
}

// OpenFile returns absolute path and content type for a file message.
func (s *Store) OpenFile(teamID, messageID string) (path, fileName, contentType string, err error) {
	tid, err1 := strconv.ParseUint(teamID, 10, 64)
	mid, err2 := strconv.ParseUint(messageID, 10, 64)
	if err1 != nil || err2 != nil {
		return "", "", "", ErrMessageNotFound
	}
	err = s.db.QueryRow(`
		SELECT COALESCE(file_path,''), COALESCE(file_name,''), COALESCE(content_type,'')
		FROM team_messages WHERE id = ? AND team_id = ? AND msg_type = ?`,
		mid, tid, MsgTypeFile,
	).Scan(&path, &fileName, &contentType)
	if err == sql.ErrNoRows || path == "" {
		return "", "", "", ErrMessageNotFound
	}
	if err != nil {
		return "", "", "", err
	}
	return filepath.Join(s.fileDir, path), fileName, contentType, nil
}

func allowedContentType(ct, fileName string) bool {
	ct = strings.ToLower(ct)
	if strings.HasPrefix(ct, "image/") {
		return true
	}
	switch ct {
	case "application/pdf", "text/plain", "application/zip",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/msword", "application/vnd.ms-excel":
		return true
	}
	lower := strings.ToLower(fileName)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".pdf", ".txt", ".zip", ".doc", ".docx", ".xls", ".xlsx"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "file"
	}
	return b.String()
}
