package middleware

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	xerrors "github.com/zeromicro/x/errors"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// JWT validates short-lived access tokens (memory / Authorization header).
func JWT(accessSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := authjwt.AccessFromHeader(r)
			if token == "" {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "missing access token"))
				return
			}
			if _, err := authjwt.Parse(accessSecret, token, authjwt.ClaimTypeAccess); err != nil {
				httpx.ErrorCtx(r.Context(), w, xerrors.New(401, "invalid or expired access token"))
				return
			}
			next(w, r)
		}
	}
}
