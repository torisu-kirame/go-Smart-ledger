package svc

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/chainstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/evmanchor"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ipfsclient"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ledgerhd"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/storage"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/txqueue"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceContext struct {
	Config    config.Config
	Ledger    *ledgersvc.Service
	Chain     chainstore.Store
	Backup    *storage.DualBackup
	IPFS      *ipfsclient.Client
	Queue     *txqueue.Queue
	EtcdLease clientv3.LeaseID
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if err := snowflake.Init(c.Snowflake.NodeID); err != nil {
		return nil, err
	}
	chain, err := chainstore.New(c.ChainStoreConfig())
	if err != nil {
		return nil, err
	}
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
	extCfg, err := evmanchor.LoadConfigFromEnv(evmanchor.Config{
		Enabled:             c.ExternalAnchor.Enabled,
		RPCURL:              c.ExternalAnchor.RPCURL,
		ChainID:             c.ExternalAnchor.ChainID,
		ChainName:           c.ExternalAnchor.ChainName,
		Contract:            c.ExternalAnchor.Contract,
		PrivateKeyHex:       c.ExternalAnchor.PrivateKeyHex,
		ExplorerURLTemplate: c.ExternalAnchor.ExplorerURLTemplate,
	})
	if err != nil {
		return nil, err
	}
	external, err := evmanchor.New(extCfg)
	if err != nil {
		return nil, err
	}

	var queue *txqueue.Queue
	if c.TxQueue.Enabled {
		queue, err = txqueue.New(chain.Submit, txqueue.Options{
			PersistPath: c.TxQueue.PersistPath,
			MaxAttempts: c.TxQueue.MaxAttempts,
			NSQ:         c.NSQ,
		})
		if err != nil {
			return nil, err
		}
	}
	return &ServiceContext{
		Config: c,
		Chain:  chain,
		Ledger: ledgersvc.New(chain, hd, queue, external),
		Backup: storage.NewDualBackup(disk, ipfs),
		IPFS:   ipfs,
		Queue:  queue,
	}, nil
}
