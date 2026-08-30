package mount

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/appconfig"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/config"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
)

// Register mounts all ledger HTTP routes onto r and starts background workers.
func Register(r router.Registrar, ac appconfig.Config) (ctx *svc.ServiceContext, lc *svc.Lifecycle, err error) {
	c := config.Config{}
	c.RestConf = ac.RestConf
	c.Chain.Backend = ac.Chain.Backend
	c.MiniLedger = ac.MiniLedger
	c.TxQueue = ac.TxQueue
	c.NSQ = ac.NSQ
	c.BackupDir = ac.BackupDir
	c.HDWallet.Mnemonic = ac.HDWallet.Mnemonic
	c.Snowflake.NodeID = ac.Snowflake.NodeID
	c.IPFS.ApiURL = ac.IPFS.ApiURL
	c.IPFS.Enabled = ac.IPFS.Enabled
	c.ExternalAnchor.Enabled = ac.ExternalAnchor.Enabled
	c.ExternalAnchor.RPCURL = ac.ExternalAnchor.RPCURL
	c.ExternalAnchor.ChainID = ac.ExternalAnchor.ChainID
	c.ExternalAnchor.ChainName = ac.ExternalAnchor.ChainName
	c.ExternalAnchor.Contract = ac.ExternalAnchor.Contract
	c.ExternalAnchor.PrivateKeyHex = ac.ExternalAnchor.PrivateKeyHex
	c.ExternalAnchor.ExplorerURLTemplate = ac.ExternalAnchor.ExplorerURLTemplate
	c.Discovery.Grpc.Enabled = false

	ctx, err = svc.NewServiceContext(c)
	if err != nil {
		return nil, nil, err
	}
	lc, err = ctx.StartBackground(c)
	if err != nil {
		return nil, nil, err
	}
	handler.RegisterHandlers(r, ctx)
	handler.RegisterExtraHandlers(r, ctx)
	handler.RegisterCollaborationHandlers(r, ctx)
	handler.RegisterTableHandlers(r, ctx)
	handler.RegisterImportAdaptiveHandlers(r, ctx)
	handler.RegisterSheetCSVImportHandlers(r, ctx)
	handler.RegisterAccountingHandlers(r, ctx)
	handler.RegisterAuditExportHandlers(r, ctx)
	handler.RegisterChainHandlers(r, ctx)
	return ctx, lc, nil
}
