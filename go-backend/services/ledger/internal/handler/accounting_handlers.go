package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

const maxAttachBytes = 15 << 20

// RegisterAccountingHandlers registers entry attachment routes only.
func RegisterAccountingHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/attachments", listAttachmentsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/attachments", uploadAttachmentHandler(serverCtx))
	r.Add(http.MethodPatch, "/api/v1/ledgers/:id/accounting/attachments/:attachId", patchAttachmentAuxHandler(serverCtx))
}

func listAttachmentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var seq uint64
		if s := r.URL.Query().Get("entrySeq"); s != "" {
			seq, _ = strconv.ParseUint(s, 10, 64)
		}
		tableID := r.URL.Query().Get("tableId")
		list, err := svcCtx.Ledger.ListAttachments(r.Context(), id, uid, tableID, seq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"attachments": list})
	}
}

func uploadAttachmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		if err := r.ParseMultipartForm(maxAttachBytes); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart"))
			return
		}
		seq, _ := strconv.ParseUint(r.FormValue("entrySeq"), 10, 64)
		if seq == 0 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "entrySeq required"))
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file required"))
			return
		}
		defer file.Close()
		body, err := io.ReadAll(io.LimitReader(file, maxAttachBytes+1))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if len(body) > maxAttachBytes {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file too large"))
			return
		}
		mime := hdr.Header.Get("Content-Type")
		tableID := r.FormValue("tableId")
		aux := &accounting.AuxiliaryDims{
			Department:   r.FormValue("department"),
			Project:      r.FormValue("project"),
			Counterparty: r.FormValue("counterparty"),
		}
		att, err := svcCtx.Ledger.LinkAttachment(r.Context(), id, uid, tableID, seq, hdr.Filename, mime, int64(len(body)), body, aux, svcCtx.Backup)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, att)
	}
}

type patchAttachmentAuxBody struct {
	Department   string `json:"department"`
	Project      string `json:"project"`
	Counterparty string `json:"counterparty"`
}

func patchAttachmentAuxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		var body patchAttachmentAuxBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		att, err := svcCtx.Ledger.UpdateAttachmentAuxiliary(r.Context(), vars["id"], uid, vars["attachId"], accounting.AuxiliaryDims(body))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, att)
	}
}
