package svc

import (
	"database/sql"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/db"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/friendstore"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/config"
)

type ServiceContext struct {
	Config  config.Config
	JWT     authjwt.Config
	Cookie  authjwt.RefreshCookieOptions
	Users   userstore.Store
	Friends *friendstore.Store
	DB      *sql.DB
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	jwtCfg := authjwt.Config{
		AccessSecret:  c.Auth.AccessSecret,
		RefreshSecret: c.Auth.RefreshSecret,
		AccessExpire:  c.Auth.AccessExpire,
		RefreshExpire: c.Auth.RefreshExpire,
	}
	cookie := authjwt.DefaultRefreshCookieOpts(int(c.Auth.RefreshExpire), c.Auth.CookieSecure)
	cookie.Domain = c.Auth.CookieDomain

	ctx := &ServiceContext{
		Config: c,
		JWT:    jwtCfg,
		Cookie: cookie,
	}

	if c.MySQL.DataSource != "" {
		sqlDB, err := db.OpenAndMigrate(c.MySQL.DataSource)
		if err != nil {
			return nil, err
		}
		mysqlUsers := userstore.NewMySQLStore(sqlDB)
		seedUser, seedPass := "admin", "admin123"
		if len(c.Users) > 0 {
			seedUser = c.Users[0].Username
			seedPass = c.Users[0].Password
		}
		if err := mysqlUsers.EnsureSeed(seedUser, seedPass); err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
		ctx.DB = sqlDB
		ctx.Users = mysqlUsers
		ctx.Friends = friendstore.New(sqlDB)
		return ctx, nil
	}

	var seed []userstore.SeedUser
	for _, u := range c.Users {
		seed = append(seed, userstore.SeedUser{Username: u.Username, Password: u.Password})
	}
	if len(seed) == 0 {
		seed = []userstore.SeedUser{{Username: "admin", Password: "admin123"}}
	}
	users, err := userstore.NewMemoryStore(seed)
	if err != nil {
		return nil, err
	}
	ctx.Users = users
	return ctx, nil
}
