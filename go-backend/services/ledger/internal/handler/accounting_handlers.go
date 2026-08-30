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

// RegisterAccountingHandlers F38–F44 accounting APIs.
func RegisterAccountingHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/chart", getChartHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/accounting/chart", putChartHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/journals", listJournalsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/journals", postJournalHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/periods", listPeriodsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/periods/:period/close", closePeriodHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/periods/:period/reopen", reopenPeriodHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/reports", reportsHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/attachments", listAttachmentsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/attachments", uploadAttachmentHandler(serverCtx))
	r.Add(http.MethodPatch, "/api/v1/ledgers/:id/accounting/attachments/:attachId", patchAttachmentAuxHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/bank-statements", listBankStatementsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/bank-statements/import", importBankHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/bank-statements/:stmtId/match", matchBankHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/budget", getBudgetHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/accounting/budget", putBudgetHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/budget/analysis", budgetAnalysisHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/aging", agingReportHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/currency", getCurrencySettingsHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/accounting/currency", putCurrencySettingsHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/currency/fx-rates", getFxRatesHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/accounting/currency/fx-rates", putFxRatesHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/currency/balances", fcBalancesHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/currency/revaluation", revaluationHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/tax/presets", taxPresetsHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/tax", getTaxTemplateHandler(serverCtx))
	r.Add(http.MethodPut, "/api/v1/ledgers/:id/accounting/tax", putTaxTemplateHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/ledgers/:id/accounting/tax/apply-preset", applyTaxPresetHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/ledgers/:id/accounting/tax/report", taxReportHandler(serverCtx))
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

func getCurrencySettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		cs, err := svcCtx.Ledger.GetCurrencySettings(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, cs)
	}
}

func putCurrencySettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var cs accounting.CurrencySettings
		if err := json.NewDecoder(r.Body).Decode(&cs); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PutCurrencySettings(r.Context(), id, uid, cs)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func getFxRatesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		if period == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "period required"))
			return
		}
		out, err := svcCtx.Ledger.GetPeriodFxRates(r.Context(), id, uid, period)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func putFxRatesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var rates accounting.PeriodFxRates
		if err := json.NewDecoder(r.Body).Decode(&rates); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PutPeriodFxRates(r.Context(), id, uid, rates)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func fcBalancesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.GetFCBalances(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"balances": list})
	}
}

func revaluationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		if period == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "period required"))
			return
		}
		rep, err := svcCtx.Ledger.GetRevaluationReport(r.Context(), id, uid, period)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, rep)
	}
}

func taxPresetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = pathvar.Vars(r)["id"]
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"presets": svcCtx.Ledger.ListTaxPresets()})
	}
}

func getTaxTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		t, err := svcCtx.Ledger.GetTaxTemplate(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, t)
	}
}

func putTaxTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var t accounting.TaxTemplate
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.PutTaxTemplate(r.Context(), id, uid, t)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

type applyTaxPresetBody struct {
	PresetID string `json:"presetId"`
}

func applyTaxPresetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var body applyTaxPresetBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		out, err := svcCtx.Ledger.ApplyTaxPreset(r.Context(), id, uid, body.PresetID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func taxReportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		period := r.URL.Query().Get("period")
		if period == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "period required"))
			return
		}
		rep, err := svcCtx.Ledger.GetTaxReport(r.Context(), id, uid, period)
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
