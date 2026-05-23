package handler

import (
	"encoding/json"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/captcha"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/friendstore"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/types"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/userinfo"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterExtraHandlers adds register, user search, and friends APIs.
func RegisterExtraHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/register", Handler: registerHandler(serverCtx)},
	}, rest.WithPrefix("/api/v1/auth"))

	registerProfileHandlers(server, serverCtx)
	registerTeamHandlers(server, serverCtx)
	registerEntryTemplateHandlers(server, serverCtx)

	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/search", Handler: userSearchHandler(serverCtx)},
	}, rest.WithPrefix("/api/v1/users"))

	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/", Handler: listFriendsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/", Handler: addFriendHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/:friendId", Handler: deleteFriendHandler(serverCtx)},
	}, rest.WithPrefix("/api/v1/friends"))
}

type registerReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

func registerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if req.Username == "" || req.Password == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "username and password required"))
			return
		}
		if len(req.Password) < 6 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "password must be at least 6 characters"))
			return
		}
		if !captcha.Verify(req.CaptchaId, req.CaptchaCode, true) {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid captcha"))
			return
		}
		user, err := svcCtx.Users.Create(req.Username, req.Password)
		if err != nil {
			if err == userstore.ErrUserExists {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(409, "username already exists"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// Default avatar: SHA-256(userID) -> go-identicon PNG on disk.
		_ = userstore.EnsureDefaultAvatar(svcCtx.AvatarDir, user.ID)
		pair, err := authjwt.Issue(svcCtx.JWT, user.ID, user.Username)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		authjwt.SetRefreshCookie(w, pair.RefreshToken, svcCtx.Cookie)
		httpx.OkJsonCtx(r.Context(), w, types.LoginResp{
			AccessToken: pair.AccessToken,
			ExpiresIn:   pair.ExpiresIn,
			TokenType:   "Bearer",
			User:        userinfo.FromStore(svcCtx, user.ID, user.Username),
		})
	}
}

func userSearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "userId query required"))
			return
		}
		user, err := svcCtx.Users.FindByID(userID)
		if err != nil {
			if err == userstore.ErrUserNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "user not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if svcCtx.Profiles != nil {
			if p, err := svcCtx.Profiles.GetProfile(userID); err == nil {
				httpx.OkJsonCtx(r.Context(), w, userinfo.ToPublic(p))
				return
			}
		}
		httpx.OkJsonCtx(r.Context(), w, userinfo.FromStore(svcCtx, user.ID, user.Username))
	}
}

func listFriendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.Friends.List(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if list == nil {
			list = []friendstore.Friend{}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"friends": list})
	}
}

type addFriendReq struct {
	FriendUserId string `json:"friendUserId"`
}

func addFriendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req addFriendReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if req.FriendUserId == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "friendUserId required"))
			return
		}
		if err := svcCtx.Friends.Add(uid, req.FriendUserId); err != nil {
			switch err {
			case friendstore.ErrCannotAddSelf:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "cannot add yourself"))
			case friendstore.ErrAlreadyFriends:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(409, "already friends"))
			case friendstore.ErrFriendNotExists:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "user not found"))
			default:
				httpx.ErrorCtx(r.Context(), w, err)
			}
			return
		}
		user, _ := svcCtx.Users.FindByID(req.FriendUserId)
		resp := types.UserInfo{Id: req.FriendUserId}
		if user != nil {
			resp.Username = user.Username
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func deleteFriendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		friendID := pathvar.Vars(r)["friendId"]
		if friendID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "friendId required"))
			return
		}
		if err := svcCtx.Friends.Remove(uid, friendID); err != nil {
			if err == friendstore.ErrFriendNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "friend not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
