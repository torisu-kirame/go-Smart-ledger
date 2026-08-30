package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/teamchat"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/teamstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

func registerTeamChatHandlers(r router.Registrar, svcCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/teams/:teamId/messages", listTeamMessagesHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams/:teamId/messages", postTeamMessageHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams/:teamId/messages/file", postTeamFileHandler(svcCtx))
	r.Add(http.MethodGet, "/api/v1/teams/:teamId/chat/files/:messageId", getTeamChatFileHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams/:teamId/read", markTeamReadHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/teams/:teamId/ledgers", addTeamLedgerHandler(svcCtx))
	r.Add(http.MethodDelete, "/api/v1/teams/:teamId/ledgers/:ledgerId", removeTeamLedgerHandler(svcCtx))
}

func markTeamReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := svcCtx.TeamChat.MarkTeamRead(teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func requireTeamMember(svcCtx *svc.ServiceContext, teamID, uid string) error {
	ok, err := svcCtx.Teams.IsMember(teamID, uid)
	if err != nil {
		return err
	}
	if !ok {
		return xerrors.New(403, "forbidden")
	}
	return nil
}

func listTeamMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		since, _ := strconv.ParseUint(r.URL.Query().Get("sinceId"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		list, err := svcCtx.TeamChat.List(teamID, since, limit)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamChatErr(err))
			return
		}
		if list == nil {
			list = []teamchat.Message{}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"messages": list})
	}
}

type postMessageBody struct {
	Body string `json:"body"`
}

func postTeamMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var body postMessageBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		msg, err := svcCtx.TeamChat.PostText(teamID, uid, body.Body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamChatErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, msg)
	}
}

func postTeamFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart form"))
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "file required"))
			return
		}
		defer file.Close()
		ct := header.Header.Get("Content-Type")
		msg, err := svcCtx.TeamChat.PostFile(teamID, uid, header.Filename, ct, file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamChatErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, msg)
	}
}

func getTeamChatFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		vars := pathvar.Vars(r)
		teamID := vars["teamId"]
		messageID := vars["messageId"]
		if err := requireTeamMember(svcCtx, teamID, uid); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		path, name, ct, err := svcCtx.TeamChat.OpenFile(teamID, messageID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamChatErr(err))
			return
		}
		if ct == "" {
			ct = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ct)
		if name != "" {
			w.Header().Set("Content-Disposition", `inline; filename="`+name+`"`)
		}
		http.ServeFile(w, r, path)
	}
}

type addLedgerBody struct {
	LedgerID string `json:"ledgerId"`
}

func addTeamLedgerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		teamID := pathvar.Vars(r)["teamId"]
		var body addLedgerBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		team, err := svcCtx.Teams.AddLedger(teamID, body.LedgerID, uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, team)
	}
}

func removeTeamLedgerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		vars := pathvar.Vars(r)
		team, err := svcCtx.Teams.RemoveLedger(vars["teamId"], vars["ledgerId"], uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, mapTeamErr(err))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, team)
	}
}

func mapTeamChatErr(err error) error {
	switch err {
	case teamchat.ErrInvalidMessage:
		return xerrors.New(400, err.Error())
	case teamchat.ErrFileTooLarge, teamchat.ErrFileType:
		return xerrors.New(400, err.Error())
	case teamchat.ErrMessageNotFound:
		return xerrors.New(404, err.Error())
	default:
		return err
	}
}

func mapTeamErr(err error) error {
	switch err {
	case teamstore.ErrTeamNotFound:
		return xerrors.New(404, err.Error())
	case teamstore.ErrUnauthorized:
		return xerrors.New(403, err.Error())
	case teamstore.ErrNeedFriend, teamstore.ErrInvalidTeam, teamstore.ErrLedgerAlreadyLinked, teamstore.ErrLedgerNotLinked, teamstore.ErrLastLedger:
		return xerrors.New(400, err.Error())
	default:
		return err
	}
}
