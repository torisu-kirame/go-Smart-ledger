package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLedgerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLedgerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLedgerLogic {
	return &GetLedgerLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetLedgerLogic) GetLedger(id, userID string) (*types.LedgerResp, error) {
	meta, err := l.svcCtx.Ledger.GetForUser(l.ctx, id, userID)
	if err != nil {
		return nil, toCodeErr(err)
	}
	return mapper.LedgerToResp(meta), nil
}
