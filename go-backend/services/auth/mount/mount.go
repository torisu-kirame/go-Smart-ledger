package mount

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/appconfig"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/config"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
)

// Register mounts all auth HTTP routes onto r.
func Register(r router.Registrar, ac appconfig.Config) (*svc.ServiceContext, error) {
	c := config.Config{}
	c.RestConf = ac.RestConf
	c.Auth = ac.Auth
	c.Database = ac.Database
	c.Avatar = ac.Avatar
	c.TeamChat = ac.TeamChat
	c.Snowflake = ac.Snowflake
	c.Users = ac.Users
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		return nil, err
	}
	handler.RegisterHandlers(r, ctx)
	handler.RegisterExtraHandlers(r, ctx)
	return ctx, nil
}
