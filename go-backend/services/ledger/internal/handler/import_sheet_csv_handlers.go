package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterSheetCSVImportHandlers one-shot CSV import into a sheet (create or append).
func RegisterSheetCSVImportHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/import/sheet-csv", importSheetCSVHandler(serverCtx))
}

type sheetCSVJSONBody struct {
	CSV        string `json:"csv"`
	TableId    string `json:"tableId,optional"`
	SheetName  string `json:"sheetName,optional"`
	SignerId   string `json:"signerId,optional"`
	AutoAnchor bool   `json:"autoAnchor,optional"`
}

func importSheetCSVHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		id := pathvar.Vars(r)["id"]
		ct := strings.ToLower(r.Header.Get("Content-Type"))

		var (
			csvData   []byte
			tableID   string
			sheetName string
			signerID  string
			autoAnchor bool
		)

		if strings.Contains(ct, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart form"))
				return
			}
			file, _, err := r.FormFile("file")
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file or csv required"))
				return
			}
			defer file.Close()
			csvData, err = io.ReadAll(file)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "read file failed"))
				return
			}
			tableID = strings.TrimSpace(r.FormValue("tableId"))
			sheetName = strings.TrimSpace(r.FormValue("sheetName"))
			signerID = strings.TrimSpace(r.FormValue("signerId"))
			autoAnchor = r.FormValue("autoAnchor") == "true" || r.FormValue("autoAnchor") == "1"
		} else {
			raw, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
				return
			}
			var body sheetCSVJSONBody
			if err := json.Unmarshal(raw, &body); err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "json body required: {csv, tableId?, sheetName?}"))
				return
			}
			csvData = []byte(body.CSV)
			tableID = strings.TrimSpace(body.TableId)
			sheetName = strings.TrimSpace(body.SheetName)
			signerID = strings.TrimSpace(body.SignerId)
			autoAnchor = body.AutoAnchor
		}

		if len(strings.TrimSpace(string(csvData))) == 0 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "csv required"))
			return
		}
		if signerID == "" {
			signerID = uid
		}

		res, err := svcCtx.Ledger.ImportSheetCSV(
			r.Context(), id, uid, signerID, csvData, tableID, sheetName, autoAnchor,
		)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		meta, _ := svcCtx.Ledger.GetForUser(r.Context(), id, uid)
		out := map[string]any{"import": res}
		if meta != nil {
			out["ledger"] = mapper.LedgerToResp(meta)
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}
