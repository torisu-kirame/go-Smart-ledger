package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/config"
)

type ServiceContext struct {
	Config config.Config
	JWT    authjwt.Config
	Cookie authjwt.RefreshCookieOptions
	Users  *userstore.MemoryStore
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
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
	jwtCfg := authjwt.Config{
		AccessSecret:  c.Auth.AccessSecret,
		RefreshSecret: c.Auth.RefreshSecret,
		AccessExpire:  c.Auth.AccessExpire,
		RefreshExpire: c.Auth.RefreshExpire,
	}
	cookie := authjwt.DefaultRefreshCookieOpts(int(c.Auth.RefreshExpire), c.Auth.CookieSecure)
	cookie.Domain = c.Auth.CookieDomain
	return &ServiceContext{
		Config: c,
		JWT:    jwtCfg,
		Cookie: cookie,
		Users:  users,
	}, nil
}
