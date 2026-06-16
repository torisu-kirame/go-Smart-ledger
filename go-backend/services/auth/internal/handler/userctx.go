package handler

import (
	"net/http"

	xerrors "github.com/zeromicro/x/errors"
)

const (
	headerUserID   = "X-User-Id"
	headerUsername = "X-Username"
)

func userIDFromRequest(r *http.Request) (string, error) {
	uid := r.Header.Get(headerUserID)
	if uid == "" {
		return "", xerrors.New(401, "unauthorized")
	}
	return uid, nil
}
