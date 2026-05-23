package logic

import (
	"context"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/userinfo"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
	w      http.ResponseWriter
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request, w http.ResponseWriter) *RefreshLogic {
	return &RefreshLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, r: r, w: w}
}

func (l *RefreshLogic) Refresh() (*types.RefreshResp, error) {
	refresh, err := authjwt.ReadRefreshCookie(l.r)
	if err != nil {
		return nil, xerrors.New(401, "refresh token required")
	}
	claims, err := authjwt.Parse(l.svcCtx.JWT.RefreshSecret, refresh, authjwt.ClaimTypeRefresh)
	if err != nil {
		authjwt.ClearRefreshCookie(l.w, l.svcCtx.Cookie)
		return nil, xerrors.New(401, "invalid refresh token")
	}
	pair, err := authjwt.RefreshAccess(l.svcCtx.JWT, refresh)
	if err != nil {
		authjwt.ClearRefreshCookie(l.w, l.svcCtx.Cookie)
		return nil, xerrors.New(401, "invalid refresh token")
	}
	authjwt.SetRefreshCookie(l.w, pair.RefreshToken, l.svcCtx.Cookie)
	return &types.RefreshResp{
		AccessToken: pair.AccessToken,
		ExpiresIn:   pair.ExpiresIn,
		TokenType:   "Bearer",
		User: userinfo.FromStore(l.svcCtx, claims.UserID, claims.Username),
	}, nil
}
