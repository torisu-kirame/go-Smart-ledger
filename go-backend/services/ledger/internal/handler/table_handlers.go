package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ledgersvc"
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
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/tables/reorder", reorderTablesHandler(serverCtx))
	r.Add(http.MethodPatch, "/api/v1/ledgers/:id/tables/:tableId", updateTableHandler(serverCtx))
	r.Add(http.MethodDelete, "/api/v1/ledgers/:id/tables/:tableId", deleteTableHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/tables/:tableId/row-order", reorderEntriesHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/tables/:tableId/sheet-edit", commitSheetEditHandler(serverCtx))
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

type reorderTablesBody struct {
	TableIds []string `json:"tableIds"`
}

func reorderTablesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body reorderTablesBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		meta, err := svcCtx.Ledger.ReorderTables(r.Context(), pathvar.Vars(r)["id"], uid, body.TableIds)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

type reorderEntriesBody struct {
	RowOrder []uint64 `json:"rowOrder"`
}

func reorderEntriesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body reorderEntriesBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		vars := pathvar.Vars(r)
		meta, err := svcCtx.Ledger.ReorderEntries(r.Context(), vars["id"], uid, vars["tableId"], body.RowOrder)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

type sheetEditBody struct {
	EntrySchema *types.EntrySchemaReq `json:"entrySchema,optional"`
	VoidSeqs    []uint64              `json:"voidSeqs,optional"`
	NewRows     []map[string]any      `json:"newRows,optional"`
	RowOrder    []uint64              `json:"rowOrder,optional"`
	SignerId    string                `json:"signerId,optional"`
}

func commitSheetEditHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
			return
		}
		var body sheetEditBody
		if err := json.Unmarshal(raw, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json body"))
			return
		}
		vars := pathvar.Vars(r)
		edit := ledgersvc.SheetEditCommit{
			TableID:  vars["tableId"],
			VoidSeqs: body.VoidSeqs,
			NewRows:  body.NewRows,
			RowOrder: body.RowOrder,
			SignerID: body.SignerId,
		}
		if body.EntrySchema != nil {
			s := mapper.EntrySchemaFromReq(*body.EntrySchema)
			edit.EntrySchema = &s
		}
		meta, err := svcCtx.Ledger.CommitSheetEdit(r.Context(), vars["id"], uid, edit)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}
