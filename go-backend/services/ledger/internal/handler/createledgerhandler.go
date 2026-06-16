// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func createLedgerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateLedgerReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if req.CreatorId == "" {
			req.CreatorId = r.Header.Get("X-User-Id")
		}

		l := logic.NewCreateLedgerLogic(r.Context(), svcCtx)
		resp, err := l.CreateLedger(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
