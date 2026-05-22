package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListLedgersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLedgersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLedgersLogic {
	return &ListLedgersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListLedgersLogic) ListLedgers() ([]types.LedgerResp, error) {
	list, err := l.svcCtx.Ledger.List(l.ctx)
	if err != nil {
		return nil, toCodeErr(err)
	}
	out := make([]types.LedgerResp, len(list))
	for i, m := range list {
		out[i] = *mapper.LedgerToResp(m)
	}
	return out, nil
}
