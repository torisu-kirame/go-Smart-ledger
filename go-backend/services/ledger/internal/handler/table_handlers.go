package handler

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterTableHandlers F49 ledger multi-table APIs.
func RegisterTableHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodPatch, "/api/v1/ledgers/:id/multi-table", setMultiTableHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/tables", createTableHandler(serverCtx))
	r.Add(http.MethodPatch, "/api/v1/ledgers/:id/tables/:tableId", updateTableHandler(serverCtx))
	r.Add(http.MethodDelete, "/api/v1/ledgers/:id/tables/:tableId", deleteTableHandler(serverCtx))
}

type multiTableBody struct {
	Enabled bool `json:"enabled"`
}

func setMultiTableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body multiTableBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		meta, err := svcCtx.Ledger.SetMultiTableEnabled(r.Context(), pathvar.Vars(r)["id"], uid, body.Enabled)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

type createTableBody struct {
	Name        string              `json:"name"`
	EntrySchema types.EntrySchemaReq `json:"entrySchema"`
}

func createTableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body createTableBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		schema := mapper.EntrySchemaFromReq(body.EntrySchema)
		meta, err := svcCtx.Ledger.CreateTable(r.Context(), pathvar.Vars(r)["id"], uid, body.Name, schema)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

type updateTableBody struct {
	Name        string               `json:"name,optional"`
	EntrySchema *types.EntrySchemaReq `json:"entrySchema,optional"`
}

func updateTableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body updateTableBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var schema *domain.EntrySchema
		if body.EntrySchema != nil {
			s := mapper.EntrySchemaFromReq(*body.EntrySchema)
			schema = &s
		}
		vars := pathvar.Vars(r)
		meta, err := svcCtx.Ledger.UpdateTable(r.Context(), vars["id"], uid, vars["tableId"], body.Name, schema)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

func deleteTableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		vars := pathvar.Vars(r)
		meta, err := svcCtx.Ledger.DeleteTable(r.Context(), vars["id"], uid, vars["tableId"])
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}
