package teamstore

import (
	"sort"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/teamchat"
)

// TeamInbox is a team row for the list UI with chat summary.
type TeamInbox struct {
	Team
	LastMessage  *teamchat.LastMessagePreview `json:"lastMessage,omitempty"`
	UnreadCount  int                          `json:"unreadCount"`
	LastActiveAt time.Time                    `json:"lastActiveAt"`
}

// ListInboxByUser returns teams enriched with last message and unread counts.
func (s *Store) ListInboxByUser(userID string, chat *teamchat.Store) ([]TeamInbox, error) {
	teams, err := s.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]TeamInbox, 0, len(teams))
	for _, t := range teams {
		item := TeamInbox{Team: t, LastActiveAt: t.CreatedAt}
		if chat != nil {
			if last, err := chat.LastMessage(t.ID); err == nil && last != nil {
				item.LastMessage = last
				item.LastActiveAt = last.CreatedAt
			}
			if n, err := chat.UnreadCount(t.ID, userID); err == nil {
				item.UnreadCount = n
			}
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastActiveAt.After(out[j].LastActiveAt)
	})
	return out, nil
}
