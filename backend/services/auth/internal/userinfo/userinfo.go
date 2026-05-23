package userinfo

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/types"
)

// FromStore builds API user info with profile fields when available.
func FromStore(svcCtx *svc.ServiceContext, id, username string) types.UserInfo {
	info := types.UserInfo{Id: id, Username: username, AvatarUrl: userstore.AvatarPath(id)}
	if svcCtx.Profiles != nil {
		if p, err := svcCtx.Profiles.GetProfile(id); err == nil {
			return FromProfile(p)
		}
	}
	return info
}

// FromProfile maps store profile to UserInfo.
func FromProfile(p *userstore.Profile) types.UserInfo {
	avatar := p.AvatarURL
	if avatar == "" {
		avatar = userstore.AvatarPath(p.ID)
	}
	return types.UserInfo{
		Id:        p.ID,
		Username:  p.Username,
		Nickname:  p.Nickname,
		AvatarUrl: avatar,
		CreatedAt: p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// ToPublic maps store profile to public API shape.
func ToPublic(p *userstore.Profile) types.UserPublicProfile {
	avatar := p.AvatarURL
	if avatar == "" {
		avatar = userstore.AvatarPath(p.ID)
	}
	return types.UserPublicProfile{
		Id:        p.ID,
		Username:  p.Username,
		Nickname:  p.Nickname,
		AvatarUrl: avatar,
		CreatedAt: p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
