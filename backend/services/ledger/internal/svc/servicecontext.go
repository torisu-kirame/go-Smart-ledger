package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/storage"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Ledger *ledgersvc.Service
	Chain  *miniledgerclient.Client
	Backup *storage.DiskBackup
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	chain := miniledgerclient.New(c.MiniLedger.BaseURL)
	b, err := storage.NewDiskBackup(c.BackupDir)
	if err != nil {
		return nil, err
	}
	return &ServiceContext{
		Config: c,
		Chain:  chain,
		Ledger: ledgersvc.New(chain),
		Backup: b,
	}, nil
}
