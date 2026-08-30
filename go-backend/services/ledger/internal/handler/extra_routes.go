package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importfile"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
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
		{Method: http.MethodPost, Path: "/ledgers/:id/restore/commit", Handler: restoreCommitHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/rag-export", Handler: ragExportHandler(serverCtx)},
	}, prefix)
}

func ragExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		id := pathvar.Vars(r)["id"]
		out, err := svcCtx.Ledger.ExportRAG(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
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
		var data []byte
		var err error
		if lid := r.URL.Query().Get("ledgerId"); lid != "" {
			if meta, errGet := svcCtx.Ledger.Get(r.Context(), lid); errGet == nil {
				domain.NormalizeLedgerTables(meta)
				if meta.MultiTableEnabled && len(meta.Tables) > 0 {
					data, err = importxlsx.BuildTemplateWorkbook(meta.Tables)
				} else {
					schema = domain.ResolveEntrySchema(meta.EntrySchema)
					data, err = importxlsx.BuildTemplate(schema)
				}
			}
		}
		if data == nil {
			if tid := r.URL.Query().Get("templateId"); tid != "" {
				for _, t := range domain.BuiltinTemplates() {
					if t.TemplateID == tid {
						schema = t
						break
					}
				}
			}
			data, err = importxlsx.BuildTemplate(schema)
		}
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
		file, hdr, err := r.FormFile("file")
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
		domain.NormalizeLedgerTables(meta)
		tableID := r.FormValue("tableId")
		if tableID == "" {
			tableID = r.URL.Query().Get("tableId")
		}
		if meta.MultiTableEnabled && tableID == "" {
			sheets, err := importxlsx.ParseForTables(data, meta.Tables)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
				return
			}
			httpx.OkJsonCtx(r.Context(), w, map[string]any{
				"multiTable": true,
				"sheets":     sheets,
			})
			return
		}
		schema, err := domain.SchemaForTable(meta, tableID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		sheetName := ""
		if t := domain.TableByID(meta, tableID); t != nil {
			sheetName = t.Name
		}
		rows, err := importfile.ParseWithSchema(data, hdr.Filename, sheetName, schema)
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
			"multiTable":  meta.MultiTableEnabled,
			"tableId":     domain.ResolveTableID(meta, tableID),
			"rows":        rows,
			"valid":       valid,
			"invalid":     invalid,
			"total":       len(rows),
			"entrySchema": schema,
		})
	}
}

type importCommitBody struct {
	SignerID       string                  `json:"signerId"`
	TableId        string                  `json:"tableId,optional"`
	Rows           []importxlsx.RowPreview `json:"rows"`
	AutoAnchor     bool                    `json:"autoAnchor,optional"`
	AutoBackup     bool                    `json:"autoBackup,optional"`
	BackupPassword string                  `json:"backupPassword,optional"`
}

func importCommitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		var body importCommitBody
		raw, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid body"))
			return
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json body"))
			return
		}
		if body.SignerID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "signerId required"))
			return
		}
		res, err := svcCtx.Ledger.BatchImport(r.Context(), id, body.SignerID, body.TableId, body.Rows, body.AutoAnchor)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		if body.AutoBackup && body.BackupPassword != "" {
			bk, err := createBackup(r, svcCtx, id, body.BackupPassword, body.SignerID)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, map[string]any{
					"import":      res,
					"backupError": err.Error(),
				})
				return
			}
			if ref, ok := bk["ref"].(string); ok {
				res.BackupRef = ref
			}
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
		result, err := createBackup(r, svcCtx, id, body.Password, r.Header.Get("X-User-Id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(500, err.Error()))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, result)
	}
}

type restoreBody struct {
	Ref       string `json:"ref"`
	Password  string `json:"password"`
	IPFSCID   string `json:"ipfsCid,optional"`
	Overwrite bool   `json:"overwrite,optional"`
}

func restorePreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = pathvar.Vars(r)["id"]
		var body restoreBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		snap, err := loadSnapshotFromBackup(r, svcCtx, body)
		if err != nil {
			writeRestoreErr(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, snap)
	}
}

func restoreCommitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ledgerID := pathvar.Vars(r)["id"]
		var body restoreBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		snap, err := loadSnapshotFromBackup(r, svcCtx, body)
		if err != nil {
			writeRestoreErr(w, r, err)
			return
		}
		signer := r.Header.Get("X-User-Id")
		if err := svcCtx.Ledger.RestoreSnapshot(r.Context(), ledgerID, snap, ledgersvc.RestoreOptions{
			Overwrite: body.Overwrite,
			SignerID:  signer,
		}); err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		meta, _ := svcCtx.Ledger.Get(r.Context(), ledgerID)
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"ledgerId":   ledgerID,
			"restored":   true,
			"latestSeq":  meta.LatestSeq,
			"latestRoot": meta.LatestRoot,
		})
	}
}

func loadSnapshotFromBackup(r *http.Request, svcCtx *svc.ServiceContext, body restoreBody) (*ledgersvc.LedgerSnapshot, error) {
	if body.Ref == "" || body.Password == "" {
		return nil, xerrors.New(400, "ref and password required")
	}
	plain, err := svcCtx.Backup.Get(r.Context(), body.Ref, body.Password, body.IPFSCID)
	if err != nil {
		return nil, xerrors.New(404, "备份不存在或密码错误")
	}
	var snap ledgersvc.LedgerSnapshot
	if err := json.Unmarshal(plain, &snap); err != nil {
		return nil, xerrors.New(500, "invalid backup payload")
	}
	return &snap, nil
}

func writeRestoreErr(w http.ResponseWriter, r *http.Request, err error) {
	if xe, ok := err.(*xerrors.CodeMsg); ok {
		httpx.ErrorCtx(r.Context(), w, xe)
		return
	}
	httpx.ErrorCtx(r.Context(), w, err)
}

func createBackup(r *http.Request, svcCtx *svc.ServiceContext, ledgerID, password, signerID string) (map[string]any, error) {
	_, raw, err := svcCtx.Ledger.BuildSnapshot(r.Context(), ledgerID)
	if err != nil {
		return nil, err
	}
	putRes, err := svcCtx.Backup.Put(r.Context(), ledgerID, password, raw)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"ref":     putRes.Ref,
		"ipfsCid": putRes.IPFSCID,
	}
	if signerID == "" {
		return out, nil
	}
	anchor, err := svcCtx.Ledger.RecordBackupAnchor(r.Context(), ledgerID, signerID, putRes.Ref, putRes.IPFSCID)
	if err != nil {
		out["anchorError"] = err.Error()
		return out, nil
	}
	out["anchoredOnChain"] = anchor.Anchored
	out["anchorSeq"] = anchor.Seq
	return out, nil
}
