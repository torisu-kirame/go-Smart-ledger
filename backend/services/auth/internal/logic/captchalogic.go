package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/captcha"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/auth/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CaptchaLogic) Captcha() (*types.CaptchaResp, error) {
	id, b64, err := captcha.Generate()
	if err != nil {
		return nil, xerrors.New(500, "captcha generate failed")
	}
	return &types.CaptchaResp{CaptchaId: id, Image: b64}, nil
}
