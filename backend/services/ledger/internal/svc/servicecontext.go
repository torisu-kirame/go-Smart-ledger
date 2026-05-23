package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ipfsclient"
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
	Backup *storage.DualBackup
	IPFS   *ipfsclient.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if err := snowflake.Init(c.Snowflake.NodeID); err != nil {
		return nil, err
	}
	chain := miniledgerclient.New(c.MiniLedger.BaseURL)
	disk, err := storage.NewDiskBackup(c.BackupDir)
	if err != nil {
		return nil, err
	}
	var ipfs *ipfsclient.Client
	if c.IPFS.Enabled && c.IPFS.ApiURL != "" {
		ipfs = ipfsclient.New(c.IPFS.ApiURL)
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
		Backup: storage.NewDualBackup(disk, ipfs),
		IPFS:   ipfs,
	}, nil
}
