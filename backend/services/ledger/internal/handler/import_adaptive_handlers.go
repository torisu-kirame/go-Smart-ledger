package handler

import (
	"io"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importfile"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterImportAdaptiveHandlers ledger import with adaptive schema / new table.
func RegisterImportAdaptiveHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/ledgers/:id/import/adaptive/preview", Handler: importAdaptivePreviewHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/import/adaptive/commit", Handler: importAdaptiveCommitHandler(serverCtx)},
	}, prefix)
}

func importAdaptivePreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
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
		if domain.IsProfessionalBookkeeping(meta) {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "adaptive import only for simple ledgers"))
			return
		}
		result, err := importfile.ParseAdaptive(data, hdr.Filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"adaptive":          true,
			"proposedTableName": result.ProposedTableName,
			"entrySchema":       result.EntrySchema,
			"rows":              result.Rows,
			"valid":             result.Valid,
			"invalid":           result.Invalid,
			"total":             result.Total,
			"fileKind":          result.FileKind,
			"willEnableMultiTable": !meta.MultiTableEnabled,
		})
	}
}

type adaptiveCommitBody struct {
	SignerID       string                  `json:"signerId"`
	TableName      string                  `json:"tableName,optional"`
	EntrySchema    domain.EntrySchema      `json:"entrySchema"`
	Rows           []importxlsx.RowPreview `json:"rows"`
	AutoAnchor     bool                    `json:"autoAnchor"`
	AutoBackup     bool                    `json:"autoBackup"`
	BackupPassword string                  `json:"backupPassword,omitempty"`
}

func importAdaptiveCommitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body adaptiveCommitBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if body.SignerID == "" {
			body.SignerID = uid
		}
		res, err := svcCtx.Ledger.ImportAdaptiveCommit(
			r.Context(), id, uid, body.SignerID,
			body.TableName, body.EntrySchema, body.Rows, body.AutoAnchor,
		)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		if body.AutoBackup && body.BackupPassword != "" {
			bk, err := createBackup(r, svcCtx, id, body.BackupPassword, body.SignerID)
			if err == nil {
				if ref, ok := bk["ref"].(string); ok {
					res.Import.BackupRef = ref
				}
			}
		}
		meta, _ := svcCtx.Ledger.Get(r.Context(), id)
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"adaptive": res,
			"ledger":   mapper.LedgerToResp(meta),
		})
	}
}
