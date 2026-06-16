// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

func listEventsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListEventsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		id := pathvar.Vars(r)["id"]
		l := logic.NewListEventsLogic(r.Context(), svcCtx)
		resp, err := l.ListEvents(id, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
