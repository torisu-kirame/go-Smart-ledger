package handler

import (
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/logic"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterCollaborationHandlers F17/F18/F19 APIs.
func RegisterCollaborationHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	prefix := rest.WithPrefix("/api/v1")
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/ledgers/invites/mine", Handler: myInvitesHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/pending", Handler: listPendingHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/entries/propose", Handler: proposeEntryHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/pending/:pendingId/approve", Handler: approvePendingHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/pending/:pendingId/reject", Handler: rejectPendingHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/members/invite", Handler: inviteMemberHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/invites", Handler: listLedgerInvitesHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/invites/accept", Handler: acceptInviteHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/ledgers/:id/sync", Handler: syncLedgerHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/ledgers/:id/encryption/rotate", Handler: rotateKeysHandler(serverCtx)},
	}, prefix)
}

func userIDFromHeader(r *http.Request) string {
	return r.Header.Get("X-User-Id")
}

func myInvitesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		list, err := svcCtx.Ledger.ListInvitesForUser(r.Context(), uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"invites": list})
	}
}

func listPendingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.ListPending(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"pending": list})
	}
}

type proposeBody struct {
	Entry typesEntry `json:"entry"`
}

type typesEntry struct {
	SignerId string            `json:"signerId,optional"`
	SchemaId string            `json:"schemaId,optional"`
	Data     map[string]string `json:"data,optional"`
	Date     string            `json:"date,optional"`
	Type     string            `json:"type,optional"`
	Amount   string            `json:"amount,optional"`
	Category string            `json:"category,optional"`
	Note     string            `json:"note,optional"`
}

func proposeEntryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		var body proposeBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		signer := body.Entry.SignerId
		if signer == "" {
			signer = uid
		}
		entry := domain.EntryPayload{
			SchemaID: body.Entry.SchemaId,
			Data:     body.Entry.Data,
			Date:     body.Entry.Date,
			Type:     body.Entry.Type,
			Amount:   body.Entry.Amount,
			Category: body.Entry.Category,
			Note:     body.Entry.Note,
		}
		pending, ev, err := svcCtx.Ledger.ProposeEntry(r.Context(), id, signer, entry)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		out := map[string]any{}
		if pending != nil {
			out["pending"] = pending
			out["status"] = "pending_approval"
		}
		if ev != nil {
			out["event"] = eventToMap(ev)
			out["status"] = "committed"
		}
		httpx.OkJsonCtx(r.Context(), w, out)
	}
}

func approvePendingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		ev, err := svcCtx.Ledger.ApprovePending(r.Context(), vars["id"], vars["pendingId"], uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		if ev == nil {
			httpx.OkJsonCtx(r.Context(), w, map[string]any{"status": "awaiting_more_approvals"})
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"status": "committed", "event": eventToMap(ev)})
	}
}

func rejectPendingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		uid := userIDFromHeader(r)
		if uid == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "unauthorized"))
			return
		}
		if err := svcCtx.Ledger.RejectPending(r.Context(), vars["id"], vars["pendingId"], uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type inviteBody struct {
	InviteeId string `json:"inviteeId"`
	Role      string `json:"role,optional"`
}

func inviteMemberHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var body inviteBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		inv, err := svcCtx.Ledger.InviteMember(r.Context(), id, uid, body.InviteeId, body.Role)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, inv)
	}
}

func listLedgerInvitesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		list, err := svcCtx.Ledger.ListInvites(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"invites": list})
	}
}

func acceptInviteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		meta, err := svcCtx.Ledger.AcceptInvite(r.Context(), id, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

func syncLedgerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		since, _ := strconv.ParseUint(r.URL.Query().Get("sinceSeq"), 10, 64)
		events, meta, err := svcCtx.Ledger.SyncEvents(r.Context(), id, uid, since)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		evOut := make([]map[string]any, len(events))
		for i, e := range events {
			evOut[i] = eventToMap(&e)
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"ledger": mapper.LedgerToResp(meta),
			"events": evOut,
			"sinceSeq": since,
		})
	}
}

type rotateBody struct {
	WrappedKeys map[string]string `json:"wrappedKeys"`
}

func rotateKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathvar.Vars(r)["id"]
		uid := userIDFromHeader(r)
		var body rotateBody
		if err := httpx.Parse(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		meta, err := svcCtx.Ledger.RotateGroupKeys(r.Context(), id, uid, body.WrappedKeys)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.ToCodeErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, mapper.LedgerToResp(meta))
	}
}

func eventToMap(ev *domain.EventRecord) map[string]any {
	if ev == nil {
		return nil
	}
	return map[string]any{
		"seq":       ev.Seq,
		"type":      ev.Type,
		"hash":      ev.Hash,
		"signerId":  ev.SignerID,
		"createdAt": ev.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
