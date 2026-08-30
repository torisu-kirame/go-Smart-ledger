package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/registry"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RegisterDiscoveryHandlers exposes etcd service discovery (F28).
func RegisterDiscoveryHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	if !serverCtx.Config.Discovery.Etcd.Enabled {
		return
	}
	r.Add(http.MethodGet, "/api/v1/discovery/services", discoveryListHandler(serverCtx))
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
