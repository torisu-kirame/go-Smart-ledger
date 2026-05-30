package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/aiproxy"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterAIHandlers proxies chat via OpenClaw Gateway (F34 assistant UI).
func RegisterAIHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/ai/chat", Handler: aiChatHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ai/test", Handler: aiTestHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ai/agent/load", Handler: aiAgentLoadHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ai/agent/save", Handler: aiAgentSaveHandler(serverCtx)},
	}, prefix)
}

func aiChatHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userIDFromHeader(r) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var req aiproxy.OpenClawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if err := aiproxy.ProxyOpenClawChat(w, r, req); err != nil {
			if err == aiproxy.ErrInvalidBaseURL {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
	}
}

func aiTestHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userIDFromHeader(r) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var req aiproxy.TestOpenClawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if err := aiproxy.TestOpenClaw(r, req); err != nil {
			if err == aiproxy.ErrInvalidBaseURL || err == aiproxy.ErrOpenClawGateway || err == aiproxy.ErrOpenClawTest {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"ok": true})
	}
}

func aiAgentLoadHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userIDFromHeader(r) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var req aiproxy.AgentStorageLoadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if strings.TrimSpace(req.AgentPath) == "" || strings.TrimSpace(req.ChatHistoryPath) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "agentPath and chatHistoryPath required"))
			return
		}
		out, err := aiproxy.LoadAgentStorage(req)
		if err != nil {
			if err == aiproxy.ErrAgentPathInvalid {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func aiAgentSaveHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userIDFromHeader(r) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var req aiproxy.AgentStorageSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if strings.TrimSpace(req.AgentPath) == "" || strings.TrimSpace(req.ChatHistoryPath) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "agentPath and chatHistoryPath required"))
			return
		}
		if len(req.Messages) == 0 && len(req.WorkspaceFiles) == 0 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "nothing to save"))
			return
		}
		if err := aiproxy.SaveAgentStorage(req); err != nil {
			if err == aiproxy.ErrAgentPathInvalid {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"ok": true})
	}
}
