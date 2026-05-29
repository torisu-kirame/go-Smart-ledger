package handler

import (
	"encoding/json"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/aiproxy"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterAIHandlers proxies OpenAI-compatible chat to local LLM (F34 assistant UI).
func RegisterAIHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/ai/chat", Handler: aiChatHandler(serverCtx)},
	}, prefix)
}

func aiChatHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userIDFromHeader(r) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var req aiproxy.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if err := aiproxy.ProxyChat(w, r, req); err != nil {
			if err == aiproxy.ErrInvalidBaseURL {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
	}
}
