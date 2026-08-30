package handler

import (
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterAuditExportHandlers F48 audit package export.
func RegisterAuditExportHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/audit-export", auditExportHandler(serverCtx))
}

func auditExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		id := pathvar.Vars(r)["id"]
		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "xlsx"
		}
		if format == "excel" {
			format = "xlsx"
		}
		data, mime, filename, err := svcCtx.Ledger.ExportAudit(r.Context(), id, uid, format)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		_, _ = w.Write(data)
	}
}
