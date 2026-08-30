package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importfile"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/importxlsx"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterImportAdaptiveHandlers ledger import with adaptive schema / new table.
func RegisterImportAdaptiveHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/import/adaptive/preview", importAdaptivePreviewHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/import/adaptive/commit", importAdaptiveCommitHandler(serverCtx))
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
	TableId        string                  `json:"tableId,optional"` // set → append to existing sheet
	TableName      string                  `json:"tableName,optional"` // used when creating a new sheet
	EntrySchema    domain.EntrySchema      `json:"entrySchema"`
	Rows           []importxlsx.RowPreview `json:"rows"`
	AutoAnchor     bool                    `json:"autoAnchor,optional"`
	AutoBackup     bool                    `json:"autoBackup,optional"`
	BackupPassword string                  `json:"backupPassword,optional"`
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
			body.SignerID = uid
		}
		res, err := svcCtx.Ledger.ImportAdaptiveCommit(
			r.Context(), id, uid, body.SignerID,
			body.TableId, body.TableName, body.EntrySchema, body.Rows, body.AutoAnchor,
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
