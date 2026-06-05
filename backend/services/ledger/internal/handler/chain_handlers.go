package handler

import (
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

// RegisterChainHandlers exposes chain explorer proxy and submit retry queue (F23).
func RegisterChainHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/chain/status", Handler: chainStatusHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/chain/queue", Handler: chainQueueHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/chain/queue/:id/retry", Handler: chainQueueRetryHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/chain/blocks", Handler: chainProxyHandler(serverCtx, "/blocks")},
		{Method: http.MethodGet, Path: "/chain/blocks/latest", Handler: chainProxyHandler(serverCtx, "/blocks/latest")},
		{Method: http.MethodGet, Path: "/chain/blocks/:height", Handler: chainBlockByHeightHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/chain/tx/recent", Handler: chainProxyHandler(serverCtx, "/tx/recent")},
		{Method: http.MethodGet, Path: "/chain/consensus", Handler: chainProxyHandler(serverCtx, "/consensus")},
		{Method: http.MethodGet, Path: "/chain/peers", Handler: chainProxyHandler(serverCtx, "/peers")},
	}, prefix)
}

func chainStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st, err := svcCtx.Chain.Status(r.Context())
		pending, failed := svcCtx.Ledger.ChainQueueStats()
		out := map[string]any{
			"online":         err == nil,
			"backend":        string(svcCtx.Chain.Backend()),
			"baseUrl":        svcCtx.Chain.BaseURL(),
			"queuePending":   pending,
			"queueFailed":    failed,
			"grpcEnabled":    svcCtx.Config.Discovery.Grpc.Enabled,
			"etcdRegistered": svcCtx.EtcdLease != 0,
		}
		if err == nil && st != nil {
			out["height"] = st.Height
			out["uptime"] = st.Uptime
			out["role"] = st.Role
			if st.ExplorerURL != "" {
				out["explorerUrl"] = st.ExplorerURL
			}
			if st.Backend != "" {
				out["chainBackend"] = st.Backend
			}
		} else if err != nil {
			out["error"] = err.Error()
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func chainQueueHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items := svcCtx.Ledger.ChainQueue()
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"items": items})
	}
}

func chainQueueRetryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		if err := svcCtx.Ledger.RetryChainItem(r.Context(), id); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"id": id, "retried": true})
	}
}

func chainProxyHandler(svcCtx *svc.ServiceContext, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		raw, err := svcCtx.Chain.GetRaw(r.Context(), path)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}
}

func chainBlockByHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := pathvar.Vars(r)["height"]
		if _, err := strconv.ParseUint(h, 10, 64); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		raw, err := svcCtx.Chain.GetRaw(r.Context(), "/blocks/"+h)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}
}
