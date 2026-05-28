package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

const maxAttachBytes = 15 << 20

// RegisterAccountingHandlers F38–F44 accounting APIs.
func RegisterAccountingHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/chart", Handler: getChartHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/ledgers/:id/accounting/chart", Handler: putChartHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/journals", Handler: listJournalsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/journals", Handler: postJournalHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/periods", Handler: listPeriodsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/periods/:period/close", Handler: closePeriodHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/periods/:period/reopen", Handler: reopenPeriodHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/reports", Handler: reportsHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/attachments", Handler: listAttachmentsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/attachments", Handler: uploadAttachmentHandler(serverCtx)},
		{Method: http.MethodPatch, Path: "/ledgers/:id/accounting/attachments/:attachId", Handler: patchAttachmentAuxHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/bank-statements", Handler: listBankStatementsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/bank-statements/import", Handler: importBankHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/accounting/bank-statements/:stmtId/match", Handler: matchBankHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/budget", Handler: getBudgetHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/ledgers/:id/accounting/budget", Handler: putBudgetHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/budget/analysis", Handler: budgetAnalysisHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/accounting/aging", Handler: agingReportHandler(serverCtx)},
	}, prefix)
}

func getChartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		chart, err := svcCtx.Ledger.GetChart(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, chart)
	}
}

func putChartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var chart accounting.ChartOfAccounts
		if err := json.NewDecoder(r.Body).Decode(&chart); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PutChart(r.Context(), id, uid, chart)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func listJournalsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.ListJournals(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"journals": list})
	}
}

func postJournalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var j accounting.JournalEntry
		if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PostJournal(r.Context(), id, uid, j)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func listPeriodsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.ListPeriods(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"periods": list})
	}
}

func closePeriodHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		ps, err := svcCtx.Ledger.ClosePeriod(r.Context(), vars["id"], uid, vars["period"])
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, ps)
	}
}

func reopenPeriodHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		ps, err := svcCtx.Ledger.ReopenPeriod(r.Context(), vars["id"], uid, vars["period"])
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, ps)
	}
}

func reportsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		rep, err := svcCtx.Ledger.GetReports(r.Context(), id, uid, period)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, rep)
	}
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

func getBudgetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		if period == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "period required (YYYY-MM)"))
			return
		}
		b, err := svcCtx.Ledger.GetBudget(r.Context(), id, uid, period)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, b)
	}
}

func putBudgetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var b accounting.PeriodBudget
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PutBudget(r.Context(), id, uid, b)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func budgetAnalysisHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		if period == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "period required (YYYY-MM)"))
			return
		}
		rep, err := svcCtx.Ledger.GetBudgetAnalysis(r.Context(), id, uid, period)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, rep)
	}
}

func agingReportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		rep, err := svcCtx.Ledger.GetAgingReport(
			r.Context(), id, uid,
			r.URL.Query().Get("asOf"),
			r.URL.Query().Get("receivableAccounts"),
			r.URL.Query().Get("payableAccounts"),
		)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, rep)
	}
}

func listBankStatementsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.ListBankStatements(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"statements": list})
	}
}

func importBankHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		if err := r.ParseMultipartForm(maxAttachBytes); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart"))
			return
		}
		accountCode := r.FormValue("accountCode")
		file, hdr, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file required"))
			return
		}
		defer file.Close()
		stmt, err := svcCtx.Ledger.ImportBankStatement(r.Context(), id, uid, accountCode, file, hdr.Filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, stmt)
	}
}

type matchBankBody struct {
	LineID   string `json:"lineId"`
	EntrySeq uint64 `json:"entrySeq"`
}

func matchBankHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		var body matchBankBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		stmt, err := svcCtx.Ledger.MatchBankLine(r.Context(), vars["id"], uid, vars["stmtId"], body.LineID, body.EntrySeq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, stmt)
	}
}
