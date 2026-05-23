package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterExtraHandlers adds import, backup, template routes.
func RegisterExtraHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/entry-schema/templates", Handler: schemaTemplatesHandler()},
		{Method: http.MethodGet, Path: "/import/template", Handler: templateHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/import/preview", Handler: importPreviewHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/import/commit", Handler: importCommitHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/backup", Handler: ledgerBackupHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/restore/preview", Handler: restorePreviewHandler(serverCtx)},
	}, prefix)
}

func schemaTemplatesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates := domain.BuiltinTemplates()
		out := make([]map[string]any, len(templates))
		for i, t := range templates {
			out[i] = map[string]any{
				"templateId": t.TemplateID,
				"fields":     t.Fields,
			}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"templates": out})
	}
}

func templateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schema := domain.DefaultEntrySchema()
		if lid := r.URL.Query().Get("ledgerId"); lid != "" {
			if meta, err := svcCtx.Ledger.Get(r.Context(), lid); err == nil {
				schema = domain.ResolveEntrySchema(meta.EntrySchema)
			}
		} else if tid := r.URL.Query().Get("templateId"); tid != "" {
			for _, t := range domain.BuiltinTemplates() {
				if t.TemplateID == tid {
					schema = t
					break
				}
			}
		}
		data, err := importxlsx.BuildTemplate(schema)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="smart-ledger-import-template.xlsx"`)
		_, _ = w.Write(data)
	}
}

func importPreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart form"))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file required"))
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		id := pathvar.Vars(r)["id"]
		meta, err := svcCtx.Ledger.Get(r.Context(), id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		schema := domain.ResolveEntrySchema(meta.EntrySchema)
		rows, err := importxlsx.Parse(data, schema)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
			return
		}
		valid, invalid := 0, 0
		for _, row := range rows {
			if row.Error == "" {
				valid++
			} else {
				invalid++
			}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"rows":        rows,
			"valid":       valid,
			"invalid":     invalid,
			"total":       len(rows),
			"entrySchema": schema,
		})
	}
}

type importCommitBody struct {
	SignerID     string                `json:"signerId"`
	Rows         []importxlsx.RowPreview `json:"rows"`
	AutoAnchor   bool                  `json:"autoAnchor"`
	AutoBackup   bool                  `json:"autoBackup"`
	BackupPassword string              `json:"backupPassword,omitempty"`
}

func importCommitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		var body importCommitBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if body.SignerID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "signerId required"))
			return
		}
		res, err := svcCtx.Ledger.BatchImport(r.Context(), id, body.SignerID, body.Rows, body.AutoAnchor)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		if body.AutoBackup && body.BackupPassword != "" {
			ref, err := createBackup(r, svcCtx, id, body.BackupPassword)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, map[string]any{
					"import": res,
					"backupError": err.Error(),
				})
				return
			}
			res.BackupRef = ref
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"import": res})
	}
}

type backupBody struct {
	Password string `json:"password"`
}

func ledgerBackupHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		var body backupBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if body.Password == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "password required"))
			return
		}
		meta, err := svcCtx.Ledger.Get(r.Context(), id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		if meta.AnchorStatus != "synced" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "请先封账锚定后再备份"))
			return
		}
		ref, err := createBackup(r, svcCtx, id, body.Password)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(500, err.Error()))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]string{"ref": ref})
	}
}

type restoreBody struct {
	Ref      string `json:"ref"`
	Password string `json:"password"`
}

func restorePreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = pathvar.Vars(r)["id"]
		var body restoreBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		plain, err := svcCtx.Backup.Get(r.Context(), body.Ref, body.Password)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "备份不存在或密码错误"))
			return
		}
		var snap ledgersvc.LedgerSnapshot
		if err := json.Unmarshal(plain, &snap); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(500, "invalid backup payload"))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, snap)
	}
}

func createBackup(r *http.Request, svcCtx *svc.ServiceContext, ledgerID, password string) (string, error) {
	_, raw, err := svcCtx.Ledger.BuildSnapshot(r.Context(), ledgerID)
	if err != nil {
		return "", err
	}
	return svcCtx.Backup.Put(r.Context(), ledgerID, password, raw)
}
