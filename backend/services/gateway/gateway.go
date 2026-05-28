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
	c.ApplyEnv()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(middleware.CORS(c.Cors.AllowedOrigins))
	server.Use(middleware.RateLimit(c.Security))

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	handler.RegisterDiscoveryHandlers(server, ctx)

	authProxy := proxy.Handler(c.Upstreams.Auth)
	authJWT := middleware.JWT(c.Auth.AccessSecret)(authProxy)
	ledgerProxy := middleware.JWT(c.Auth.AccessSecret)(proxy.Handler(c.Upstreams.Ledger))
	storageProxy := middleware.JWT(c.Auth.AccessSecret)(proxy.Handler(c.Upstreams.Storage))

	// 公开：认证（refresh cookie 由 auth 服务设置）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/v1/auth/captcha", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/login", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/register", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/refresh", Handler: authProxy},
		{Method: http.MethodPost, Path: "/api/v1/auth/logout", Handler: authProxy},
		{Method: http.MethodGet, Path: "/api/v1/auth/health", Handler: authProxy},
	})

	// 用户头像（公开读取，无需 JWT）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/v1/users/:userId/avatar", Handler: authProxy},
	})

	// 需 JWT：个人资料、好友
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/v1/users/me", Handler: authJWT},
		{Method: http.MethodPatch, Path: "/api/v1/users/me", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/users/me/avatar", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/users/me/delete-account", Handler: authJWT},
		{Method: http.MethodPut, Path: "/api/v1/users/me/public-key", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/users/me/public-key", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/users/search", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/friends/requests/incoming", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/friends/requests/outgoing", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/friends/requests/:fromUserId/accept", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/friends/requests/:fromUserId/reject", Handler: authJWT},
		{Method: http.MethodDelete, Path: "/api/v1/friends/requests/:toUserId", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/friends", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/friends", Handler: authJWT},
		{Method: http.MethodDelete, Path: "/api/v1/friends/:friendId", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/teams", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams/read-all", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams/:teamId/read", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/teams/:teamId/messages", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams/:teamId/messages", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams/:teamId/messages/file", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/teams/:teamId/chat/files/:messageId", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/teams/:teamId/ledgers", Handler: authJWT},
		{Method: http.MethodDelete, Path: "/api/v1/teams/:teamId/ledgers/:ledgerId", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/teams/:teamId", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/entry-templates", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/entry-templates", Handler: authJWT},
		{Method: http.MethodGet, Path: "/api/v1/entry-templates/:templateId", Handler: authJWT},
		{Method: http.MethodPut, Path: "/api/v1/entry-templates/:templateId", Handler: authJWT},
		{Method: http.MethodDelete, Path: "/api/v1/entry-templates/:templateId", Handler: authJWT},
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
		{Method: http.MethodGet, Path: "/api/v1/entry-schema/templates", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/import/template", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/preview", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/commit", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/adaptive/preview", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/import/adaptive/commit", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/backup", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/restore/preview", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/restore/commit", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/invites/mine", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/pending", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/entries/propose", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/pending/:pendingId/approve", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/pending/:pendingId/reject", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/members/invite", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/invites", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/invites/accept", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/sync", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/rag-export", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/encryption/rotate", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/storage-location", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/approval-policy", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/multi-table", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/tables", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/tables/:tableId", Handler: ledgerProxy},
		{Method: http.MethodDelete, Path: "/api/v1/ledgers/:id/tables/:tableId", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/encryption/enable", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/encryption/passphrase-view-policy", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/encryption/passphrase-view-wrap", Handler: ledgerProxy},
		{Method: http.MethodDelete, Path: "/api/v1/ledgers/:id", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/chart", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/accounting/chart", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/journals", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/journals", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/periods", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/periods/:period/close", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/periods/:period/reopen", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/reports", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/audit-export", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/attachments", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/attachments", Handler: ledgerProxy},
		{Method: http.MethodPatch, Path: "/api/v1/ledgers/:id/accounting/attachments/:attachId", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/bank-statements", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/bank-statements/import", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/bank-statements/:stmtId/match", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/budget", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/accounting/budget", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/budget/analysis", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/aging", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/currency", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/accounting/currency", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/currency/fx-rates", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/accounting/currency/fx-rates", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/currency/balances", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/currency/revaluation", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/tax/presets", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/tax", Handler: ledgerProxy},
		{Method: http.MethodPut, Path: "/api/v1/ledgers/:id/accounting/tax", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/ledgers/:id/accounting/tax/apply-preset", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/ledgers/:id/accounting/tax/report", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/users/me/verify-password", Handler: authJWT},
		{Method: http.MethodPost, Path: "/api/v1/storage/backup", Handler: storageProxy},
		{Method: http.MethodPost, Path: "/api/v1/storage/backup/fetch", Handler: storageProxy},
		{Method: http.MethodGet, Path: "/api/v1/storage/health", Handler: storageProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/status", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/queue", Handler: ledgerProxy},
		{Method: http.MethodPost, Path: "/api/v1/chain/queue/:id/retry", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/blocks", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/blocks/latest", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/blocks/:height", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/tx/recent", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/consensus", Handler: ledgerProxy},
		{Method: http.MethodGet, Path: "/api/v1/chain/peers", Handler: ledgerProxy},
	})

	// discovery/services is registered by RegisterDiscoveryHandlers when etcd is enabled

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
