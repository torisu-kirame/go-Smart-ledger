package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/storage"
	"github.com/smart-ledger/go-smart-ledger/backend/services/storage/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Backup *storage.DiskBackup
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	b, err := storage.NewDiskBackup(c.BackupDir)
	if err != nil {
		return nil, err
	}
	return &ServiceContext{Config: c, Backup: b}, nil
}
