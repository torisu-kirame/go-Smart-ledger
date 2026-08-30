package mount

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/appconfig"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/config"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/svc"
)

// Register mounts storage HTTP routes onto r.
func Register(r router.Registrar, ac appconfig.Config) (*svc.ServiceContext, error) {
	c := config.Config{}
	c.RestConf = ac.RestConf
	c.BackupDir = ac.BackupDir
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		return nil, err
	}
	handler.RegisterHandlers(r, ctx)
	return ctx, nil
}
