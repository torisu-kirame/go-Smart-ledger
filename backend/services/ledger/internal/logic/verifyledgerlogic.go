package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyLedgerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyLedgerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyLedgerLogic {
	return &VerifyLedgerLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *VerifyLedgerLogic) VerifyLedger(id string) (*types.VerifyResp, error) {
	ok, err := l.svcCtx.Ledger.Verify(l.ctx, id)
	if err != nil {
		return nil, toCodeErr(err)
	}
	return &types.VerifyResp{Valid: ok}, nil
}
