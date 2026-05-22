package svc

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Ledger *ledgersvc.Service
	Chain  *miniledgerclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	chain := miniledgerclient.New(c.MiniLedger.BaseURL)
	return &ServiceContext{
		Config: c,
		Chain:  chain,
		Ledger: ledgersvc.New(chain),
	}
}
