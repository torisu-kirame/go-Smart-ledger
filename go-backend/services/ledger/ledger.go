// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router/restadapt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/config"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ledger-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	lc, err := ctx.StartBackground(c)
	if err != nil {
		panic(err)
	}
	defer lc.StopAll(ctx)

	r := restadapt.New(server)
	handler.RegisterHandlers(r, ctx)
	handler.RegisterExtraHandlers(r, ctx)
	handler.RegisterCollaborationHandlers(r, ctx)
	handler.RegisterTableHandlers(r, ctx)
	handler.RegisterImportAdaptiveHandlers(r, ctx)
	handler.RegisterSheetCSVImportHandlers(r, ctx)
	handler.RegisterAccountingHandlers(r, ctx)
	handler.RegisterAuditExportHandlers(r, ctx)
	handler.RegisterChainHandlers(r, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
