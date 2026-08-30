package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/teamstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

func registerTeamHandlers(r router.Registrar, svcCtx *svc.ServiceContext) {
	registerTeamChatHandlers(r, svcCtx)
	r.Add(http.MethodGet, "/api/v1/teams", listTeamsHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams", createTeamHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams/read-all", markAllTeamsReadHandler(svcCtx))
	r.Add(http.MethodGet, "/api/v1/teams/:teamId", getTeamHandler(svcCtx))
}

type createTeamReq struct {
	Name          string   `json:"name"`
	LedgerID      string   `json:"ledgerId"`
	LedgerIDs     []string `json:"ledgerIds"`
	LedgerType    string   `json:"ledgerType"`
	MemberUserIDs []string `json:"memberUserIds"`
}

func createTeamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creatorID, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req createTeamReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		ledgerIDs := req.LedgerIDs
		if len(ledgerIDs) == 0 && req.LedgerID != "" {
			ledgerIDs = []string{req.LedgerID}
		}
		if req.Name == "" || len(ledgerIDs) < 1 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "name and at least one ledgerId required"))
			return
		}
		if req.LedgerType != "multi" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "team requires a multi-person ledger"))
			return
		}
		if len(req.MemberUserIDs) < 1 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "at least one friend required"))
			return
		}
		for _, fid := range req.MemberUserIDs {
			ok, err := svcCtx.Friends.AreFriends(creatorID, fid)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			if !ok {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "member must be your friend: "+fid))
				return
			}
		}
		teamID, err := snowflake.NextString()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		team, err := svcCtx.Teams.CreateWithID(teamID, req.Name, creatorID, ledgerIDs, req.MemberUserIDs)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, team)
	}
}

func listTeamsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.Teams.ListInboxByUser(uid, svcCtx.TeamChat)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if list == nil {
			list = []teamstore.TeamInbox{}
		}
		totalUnread := 0
		for _, t := range list {
			totalUnread += t.UnreadCount
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"teams": list, "totalUnread": totalUnread})
	}
}

func markAllTeamsReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.Teams.ListByUser(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		ids := make([]string, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}
		if err := svcCtx.TeamChat.MarkAllTeamsRead(uid, ids); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getTeamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		team, err := svcCtx.Teams.GetByID(teamID)
		if err != nil {
			if err == teamstore.ErrTeamNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "team not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, team)
	}
}
