package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgerhd"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
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
	if err := snowflake.Init(c.Snowflake.NodeID); err != nil {
		return nil, err
	}
	chain := miniledgerclient.New(c.MiniLedger.BaseURL)
	b, err := storage.NewDiskBackup(c.BackupDir)
	if err != nil {
		return nil, err
	}
	var hd *ledgerhd.Deriver
	if c.HDWallet.Mnemonic != "" {
		hd, err = ledgerhd.NewFromMnemonic(c.HDWallet.Mnemonic)
		if err != nil {
			return nil, err
		}
	}
	return &ServiceContext{
		Config: c,
		Chain:  chain,
		Ledger: ledgersvc.New(chain, hd),
		Backup: b,
	}, nil
}
