// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router/restadapt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/config"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/storage-api.yaml", "the config file")

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
	handler.RegisterHandlers(restadapt.New(server), ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
