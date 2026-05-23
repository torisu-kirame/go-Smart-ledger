package userstore

import "time"

// Profile is public user profile data.
type Profile struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

// ProfileStore manages extended user profile fields (MySQL).
type ProfileStore interface {
	GetProfile(id string) (*Profile, error)
	UpdateNickname(id, nickname string) (*Profile, error)
	SetAvatarURL(id, avatarURL string) error
}

// AvatarPath returns API path for user avatar.
func AvatarPath(userID string) string {
	return "/api/v1/users/" + userID + "/avatar"
}

func displayNickname(username, nickname string) string {
	if nickname != "" {
		return nickname
	}
	return username
}
