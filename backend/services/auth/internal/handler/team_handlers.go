package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/teamstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

func registerTeamHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/", Handler: listTeamsHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/", Handler: createTeamHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/:teamId", Handler: getTeamHandler(svcCtx)},
	}, rest.WithPrefix("/api/v1/teams"))
}

type createTeamReq struct {
	Name           string   `json:"name"`
	LedgerID       string   `json:"ledgerId"`
	LedgerType     string   `json:"ledgerType"`
	MemberUserIDs  []string `json:"memberUserIds"`
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
		if req.Name == "" || req.LedgerID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "name and ledgerId required"))
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
		team, err := svcCtx.Teams.CreateWithID(teamID, req.Name, req.LedgerID, creatorID, req.MemberUserIDs)
		if err != nil {
			switch err {
			case teamstore.ErrNeedFriend, teamstore.ErrLedgerNotMulti, teamstore.ErrInvalidTeam:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
			default:
				httpx.ErrorCtx(r.Context(), w, err)
			}
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
		list, err := svcCtx.Teams.ListByUser(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if list == nil {
			list = []teamstore.Team{}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"teams": list})
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
		// creator is always member; invited friends in team_members
		if team.CreatorID != uid {
			found := false
			for _, m := range team.Members {
				if m.UserID == uid {
					found = true
					break
				}
			}
			if !found {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(403, "forbidden"))
				return
			}
		}
		httpx.OkJsonCtx(r.Context(), w, team)
	}
}
