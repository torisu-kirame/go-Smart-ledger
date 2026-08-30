package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/entrytemplatestore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

func registerEntryTemplateHandlers(r router.Registrar, svcCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/entry-templates", listEntryTemplatesHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/entry-templates", createEntryTemplateHandler(svcCtx))
	r.Add(http.MethodGet, "/api/v1/entry-templates/:templateId", getEntryTemplateHandler(svcCtx))
	r.Add(http.MethodPut, "/api/v1/entry-templates/:templateId", updateEntryTemplateHandler(svcCtx))
	r.Add(http.MethodDelete, "/api/v1/entry-templates/:templateId", deleteEntryTemplateHandler(svcCtx))
}

type entryTemplateFieldReq struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type saveEntryTemplateReq struct {
	Name   string                  `json:"name"`
	Fields []entryTemplateFieldReq `json:"fields"`
}

func fieldsFromReq(in []entryTemplateFieldReq) []domain.EntryFieldDef {
	out := make([]domain.EntryFieldDef, len(in))
	for i, f := range in {
		out[i] = domain.EntryFieldDef{
			Key:      strings.TrimSpace(f.Key),
			Label:    strings.TrimSpace(f.Label),
			Type:     domain.FieldType(f.Type),
			Required: f.Required,
		}
	}
	return out
}

func listEntryTemplatesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.EntryTemplates.ListForUser(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"templates": list})
	}
}

func getEntryTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		id := pathvar.Vars(r)["templateId"]
		t, err := svcCtx.EntryTemplates.GetByID(id, uid)
		if err != nil {
			writeTemplateErr(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, t)
	}
}

func createEntryTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req saveEntryTemplateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		id, err := snowflake.NextString()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		t, err := svcCtx.EntryTemplates.CreateWithID(id, uid, req.Name, fieldsFromReq(req.Fields))
		if err != nil {
			writeTemplateErr(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, t)
	}
}

func updateEntryTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		id := pathvar.Vars(r)["templateId"]
		var req saveEntryTemplateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		t, err := svcCtx.EntryTemplates.Update(id, uid, req.Name, fieldsFromReq(req.Fields))
		if err != nil {
			writeTemplateErr(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, t)
	}
}

func deleteEntryTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		id := pathvar.Vars(r)["templateId"]
		if err := svcCtx.EntryTemplates.Delete(id, uid); err != nil {
			writeTemplateErr(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeTemplateErr(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case entrytemplatestore.ErrNotFound:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(404, err.Error()))
	case entrytemplatestore.ErrForbidden:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(403, err.Error()))
	case entrytemplatestore.ErrBuiltinRead, entrytemplatestore.ErrInvalid, domain.ErrInvalidSchema:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
	default:
		httpx.ErrorCtx(r.Context(), w, err)
	}
}
