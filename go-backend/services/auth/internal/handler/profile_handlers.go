package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/userstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/types"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/userinfo"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

func registerProfileHandlers(r router.Registrar, svcCtx *svc.ServiceContext) {
	r.Add(http.MethodGet, "/api/v1/users/me", getMeHandler(svcCtx))
	r.Add(http.MethodPatch, "/api/v1/users/me", patchMeHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/users/me/avatar", uploadAvatarHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/users/me/delete-account", deleteAccountHandler(svcCtx))
	r.Add(http.MethodPost, "/api/v1/users/me/verify-password", verifyPasswordHandler(svcCtx))
	r.Add(http.MethodPut, "/api/v1/users/me/public-key", putPublicKeyHandler(svcCtx))
	r.Add(http.MethodGet, "/api/v1/users/me/public-key", getPublicKeyHandler(svcCtx))
	r.Add(http.MethodGet, "/api/v1/users/:userId/avatar", getAvatarHandler(svcCtx))
}

type verifyPasswordReq struct {
	Password string `json:"password"`
}

func verifyPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req verifyPasswordReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if req.Password == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "password required"))
			return
		}
		if err := svcCtx.Users.VerifyPassword(uid, req.Password); err != nil {
			if err == userstore.ErrInvalidCredentials {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "invalid password"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req types.DeleteAccountReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "username and password required"))
			return
		}
		if err := svcCtx.Accounts.DeleteAccount(uid, req.Username, req.Password); err != nil {
			switch err {
			case userstore.ErrInvalidCredentials:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "invalid username or password"))
			case userstore.ErrUsernameMismatch:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "username must match current account"))
			case userstore.ErrUserNotFound:
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "user not found"))
			default:
				httpx.ErrorCtx(r.Context(), w, err)
			}
			return
		}
		userstore.RemoveAvatarFiles(svcCtx.AvatarDir, uid)
		authjwt.ClearRefreshCookie(w, svcCtx.Cookie)
		w.WriteHeader(http.StatusNoContent)
	}
}

func getMeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		p, err := svcCtx.Profiles.GetProfile(uid)
		if err != nil {
			if err == userstore.ErrUserNotFound {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(404, "user not found"))
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, userinfo.ToPublic(p))
	}
}

func patchMeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var req types.UpdateProfileReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		nickname := strings.TrimSpace(req.Nickname)
		if nickname == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "nickname required"))
			return
		}
		if utf8.RuneCountInString(nickname) > 32 {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "nickname too long (max 32)"))
			return
		}
		p, err := svcCtx.Profiles.UpdateNickname(uid, nickname)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, userinfo.ToPublic(p))
	}
}

func uploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := r.ParseMultipartForm(4 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid multipart form"))
			return
		}
		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "avatar file required"))
			return
		}
		defer file.Close()
		ct := header.Header.Get("Content-Type")
		body := io.Reader(file)
		if ct == "" {
			var rd io.Reader
			ct, rd, err = userstore.DetectImageType(file)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid image"))
				return
			}
			body = rd
		}
		if _, err := userstore.SaveAvatar(svcCtx.AvatarDir, uid, body, ct); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, err.Error()))
			return
		}
		url := userstore.AvatarPath(uid)
		if err := svcCtx.Profiles.SetAvatarURL(uid, url); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		p, _ := svcCtx.Profiles.GetProfile(uid)
		httpx.OkJsonCtx(r.Context(), w, userinfo.ToPublic(p))
	}
}

type publicKeyBody struct {
	PublicKeyPem string `json:"publicKeyPem"`
}

func putPublicKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		var body publicKeyBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "invalid json"))
			return
		}
		if strings.TrimSpace(body.PublicKeyPem) == "" {
			httpx.ErrorCtx(r.Context(), w, xerrors.New(400, "publicKeyPem required"))
			return
		}
		if err := svcCtx.Profiles.SetPublicKey(uid, body.PublicKeyPem); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getPublicKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := userIDFromRequest(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		pem, err := svcCtx.Profiles.GetPublicKey(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]string{"publicKeyPem": pem})
	}
}

func getAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := pathvar.Vars(r)["userId"]
		if userID == "" {
			http.NotFound(w, r)
			return
		}
		path, ct, err := userstore.ResolveAvatarFile(svcCtx.AvatarDir, userID)
		if err != nil {
			if genErr := userstore.EnsureDefaultAvatar(svcCtx.AvatarDir, userID); genErr != nil {
				http.NotFound(w, r)
				return
			}
			path, ct, err = userstore.ResolveAvatarFile(svcCtx.AvatarDir, userID)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Cache-Control", "public, max-age=300")
		http.ServeFile(w, r, path)
	}
}
