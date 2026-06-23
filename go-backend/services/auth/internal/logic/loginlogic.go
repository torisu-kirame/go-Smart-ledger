package logic

import (
	"context"
	"net/http"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/captcha"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/userinfo"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	w      http.ResponseWriter
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext, w http.ResponseWriter) *LoginLogic {
	return &LoginLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, w: w}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	if req.Username == "" || req.Password == "" {
		return nil, xerrors.New(400, "username and password required")
	}
	if !captcha.Verify(req.CaptchaId, strings.TrimSpace(req.CaptchaCode), true) {
		return nil, xerrors.New(400, "invalid captcha")
	}
	user, err := l.svcCtx.Users.Authenticate(req.Username, req.Password)
	if err != nil {
		return nil, xerrors.New(401, "invalid username or password")
	}
	pair, err := authjwt.Issue(l.svcCtx.JWT, user.ID, user.Username)
	if err != nil {
		return nil, xerrors.New(500, "token issue failed")
	}
	authjwt.SetRefreshCookie(l.w, pair.RefreshToken, l.svcCtx.Cookie)
	return &types.LoginResp{
		AccessToken: pair.AccessToken,
		ExpiresIn:   pair.ExpiresIn,
		TokenType:   "Bearer",
		User: userinfo.FromStore(l.svcCtx, user.ID, user.Username),
	}, nil
}
