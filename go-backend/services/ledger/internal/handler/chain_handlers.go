package handler

import (
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

// RegisterChainHandlers exposes chain explorer proxy and submit retry queue (F23).
func RegisterChainHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/chain/status", chainStatusHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/chain/queue", chainQueueHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/chain/queue/:id/retry", chainQueueRetryHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/chain/blocks", chainProxyHandler(serverCtx, "/blocks"))
	r.Add(http.MethodGet, "/api/v1/chain/blocks/latest", chainProxyHandler(serverCtx, "/blocks/latest"))
	r.Add(http.MethodGet, "/api/v1/chain/blocks/:height", chainBlockByHeightHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/chain/tx/recent", chainProxyHandler(serverCtx, "/tx/recent"))
	r.Add(http.MethodGet, "/api/v1/chain/consensus", chainProxyHandler(serverCtx, "/consensus"))
	r.Add(http.MethodGet, "/api/v1/chain/peers", chainProxyHandler(serverCtx, "/peers"))
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
