// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func putBackupHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PutBackupReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPutBackupLogic(r.Context(), svcCtx)
		resp, err := l.PutBackup(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
