package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *HealthLogic) Health() (*types.HealthResp, error) {
	online := l.svcCtx.Ledger.Online(l.ctx)
	ipfsOnline := false
	if l.svcCtx.IPFS != nil && l.svcCtx.IPFS.Enabled() {
		ipfsOnline = l.svcCtx.IPFS.Ping(l.ctx) == nil
	}
	return &types.HealthResp{Status: "ok", MiniLedgerOnline: online, IPFSOnline: ipfsOnline}, nil
}
