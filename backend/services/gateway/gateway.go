package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/config"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/handler"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/middleware"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/proxy"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(middleware.CORS(c.Cors.AllowedOrigins))

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	authProxy := proxy.Handler(c.Upstreams.Auth)
	ledgerProxy := middleware.JWT(c.Auth.AccessSecret)(proxy.Handler(c.Upstreams.Ledger))
	storageProxy := middleware.JWT(c.Auth.AccessSecret)(proxy.Handler(c.Upstreams.Storage))

	// 公开：认证（refresh cookie 由 auth 服务设置）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/v1/auth/captcha", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/login", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/refresh", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/logout", Handler: authProxy},
		{Method: http.MethodGet, Path: "/api/v1/auth/health", Handler: authProxy},
	})

	// 需 JWT 短期令牌
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/v1/ledgers", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/entries", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/events", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/anchor", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/verify", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/import/template", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/preview", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/commit", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/backup", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/restore/preview", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/storage/backup", Handler: storageProxy},
		{Method: http.MethodPost, Path: "/api/v1/storage/backup/fetch", Handler: storageProxy},
		{Method: http.MethodGet, Path: "/api/v1/storage/health", Handler: storageProxy},
	})

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
