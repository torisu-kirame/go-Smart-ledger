package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/registry"
	"github.com/smart-ledger/go-smart-ledger/backend/services/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RegisterDiscoveryHandlers exposes etcd service discovery (F28).
func RegisterDiscoveryHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	if !serverCtx.Config.Discovery.Etcd.Enabled {
		return
	}
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/discovery/services", Handler: discoveryListHandler(serverCtx)},
	}, prefix)
}

func discoveryListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eps, err := registry.Discover(r.Context(), svcCtx.Config.Discovery.Etcd)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"services": eps})
	}
}
