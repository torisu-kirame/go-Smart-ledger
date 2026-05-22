// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func refreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewRefreshLogic(r.Context(), svcCtx, r, w)
		resp, err := l.Refresh()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
