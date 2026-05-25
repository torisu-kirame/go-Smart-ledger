package svc

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/db"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/entrytemplatestore"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/friendstore"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/teamchat"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/teamstore"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	JWT       authjwt.Config
	Cookie    authjwt.RefreshCookieOptions
	Users     userstore.Store
	Profiles  userstore.ProfileStore
	Accounts  userstore.AccountStore
	Friends   *friendstore.Store
	Teams          *teamstore.Store
	TeamChat       *teamchat.Store
	EntryTemplates *entrytemplatestore.Store
	DB             *sql.DB
	AvatarDir string
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if c.Database.DataSource == "" {
		return nil, fmt.Errorf("Database.DataSource is required (configure external MySQL in etc/auth-api.yaml)")
	}
	if err := snowflake.Init(c.Snowflake.NodeID); err != nil {
		return nil, err
	}

	jwtCfg := authjwt.Config{
		AccessSecret:  c.Auth.AccessSecret,
		RefreshSecret: c.Auth.RefreshSecret,
		AccessExpire:  c.Auth.AccessExpire,
		RefreshExpire: c.Auth.RefreshExpire,
	}
	cookie := authjwt.DefaultRefreshCookieOpts(int(c.Auth.RefreshExpire), c.Auth.CookieSecure)
	cookie.Domain = c.Auth.CookieDomain

	avatarDir := c.Avatar.Dir
	if avatarDir == "" {
		avatarDir = "data/avatars"
	}
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return nil, err
	}
	teamChatDir := c.TeamChat.Dir
	if teamChatDir == "" {
		teamChatDir = "data/teamchat"
	}
	if err := os.MkdirAll(teamChatDir, 0o755); err != nil {
		return nil, err
	}

	sqlDB, err := db.OpenAndMigrate(c.Database.DataSource)
	if err != nil {
		return nil, err
	}
	mysqlUsers := userstore.NewMySQLStore(sqlDB)
	seedUser, seedPass := "admin", "admin123"
	if len(c.Users) > 0 {
		seedUser = c.Users[0].Username
		seedPass = c.Users[0].Password
	}
	seedID, err := mysqlUsers.EnsureSeed(seedUser, seedPass)
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	if seedID != "" {
		_ = userstore.EnsureDefaultAvatar(avatarDir, seedID)
	}

	return &ServiceContext{
		Config:    c,
		JWT:       jwtCfg,
		Cookie:    cookie,
		AvatarDir: avatarDir,
		DB:        sqlDB,
		Users:     mysqlUsers,
		Profiles:  mysqlUsers,
		Accounts:  mysqlUsers,
		Friends:   friendstore.New(sqlDB),
		Teams:          teamstore.New(sqlDB),
		TeamChat:       teamchat.New(sqlDB, teamChatDir),
		EntryTemplates: entrytemplatestore.New(sqlDB),
	}, nil
}
