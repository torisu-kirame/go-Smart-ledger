package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/captcha"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/friendstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/types"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/userinfo"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// RegisterExtraHandlers adds register, user search, and friends APIs.
func RegisterExtraHandlers(r router.Registrar, serverCtx *svc.ServiceContext) {
	r.Add(http.MethodPost, "/api/v1/auth/register", registerHandler(serverCtx))

	registerProfileHandlers(r, serverCtx)
	registerTeamHandlers(r, serverCtx)
	registerEntryTemplateHandlers(r, serverCtx)

	r.Add(http.MethodGet, "/api/v1/users/search", userSearchHandler(serverCtx))

	r.Add(http.MethodGet, "/api/v1/friends/requests/incoming", listIncomingFriendRequestsHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/friends/requests/outgoing", listOutgoingFriendRequestsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/friends/requests/:fromUserId/accept", acceptFriendRequestHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/friends/requests/:fromUserId/reject", rejectFriendRequestHandler(serverCtx))
	r.Add(http.MethodDelete, "/api/v1/friends/requests/:toUserId", cancelFriendRequestHandler(serverCtx))
	r.Add(http.MethodGet, "/api/v1/friends", listFriendsHandler(serverCtx))
	r.Add(http.MethodPost, "/api/v1/friends", addFriendHandler(serverCtx))
	r.Add(http.MethodDelete, "/api/v1/friends/:friendId", deleteFriendHandler(serverCtx))
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
		req.Username = strings.TrimSpace(strings.ToLower(req.Username))
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
		if err := svcCtx.Friends.SendRequest(uid, req.FriendUserId); err != nil {
			writeFriendRequestError(r, w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"status":     friendstore.StatusPending,
			"toUserId":   req.FriendUserId,
			"fromUserId": uid,
		})
	}
}

func listIncomingFriendRequestsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.Friends.ListIncomingRequests(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if list == nil {
			list = []friendstore.FriendRequest{}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"requests": list})
	}
}

func listOutgoingFriendRequestsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list, err := svcCtx.Friends.ListOutgoingRequests(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if list == nil {
			list = []friendstore.FriendRequest{}
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"requests": list})
	}
}

func acceptFriendRequestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		fromID := pathvar.Vars(r)["fromUserId"]
		if fromID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "fromUserId required"))
			return
		}
		if err := svcCtx.Friends.AcceptRequest(uid, fromID); err != nil {
			if err == friendstore.ErrRequestNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "friend request not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func rejectFriendRequestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		fromID := pathvar.Vars(r)["fromUserId"]
		if fromID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "fromUserId required"))
			return
		}
		if err := svcCtx.Friends.RejectRequest(uid, fromID); err != nil {
			if err == friendstore.ErrRequestNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "friend request not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func cancelFriendRequestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		toID := pathvar.Vars(r)["toUserId"]
		if toID == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "toUserId required"))
			return
		}
		if err := svcCtx.Friends.CancelRequest(uid, toID); err != nil {
			if err == friendstore.ErrRequestNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "friend request not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeFriendRequestError(r *http.Request, w http.ResponseWriter, err error) {
	switch err {
	case friendstore.ErrCannotAddSelf:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "cannot add yourself"))
	case friendstore.ErrAlreadyFriends:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(409, "already friends"))
	case friendstore.ErrFriendNotExists:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "user not found"))
	case friendstore.ErrRequestPending:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(409, "friend request already sent"))
	case friendstore.ErrIncomingPending:
		httpx.ErrorCtx(r.Context(), w, xerrors.New(409, "incoming friend request pending, accept it instead"))
	default:
		httpx.ErrorCtx(r.Context(), w, err)
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
