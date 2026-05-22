package logic

import (
	"context"
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	w      http.ResponseWriter
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext, w http.ResponseWriter) *LogoutLogic {
	return &LogoutLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, w: w}
}

func (l *LogoutLogic) Logout() error {
	authjwt.ClearRefreshCookie(l.w, l.svcCtx.Cookie)
	return nil
}
