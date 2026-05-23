package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/mapper"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLedgerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLedgerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLedgerLogic {
	return &CreateLedgerLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateLedgerLogic) CreateLedger(req *types.CreateLedgerReq) (*types.LedgerResp, error) {
	lt, err := mapper.ParseLedgerType(req.Type)
	if err != nil {
		return nil, toCodeErr(err)
	}
	meta, err := l.svcCtx.Ledger.Create(l.ctx, lt, req.Name, req.CreatorId, mapper.MembersFromReq(req.Members), mapper.EntrySchemaFromReq(req.EntrySchema), mapper.CreateOptionsFromReq(req))
	if err != nil {
		return nil, toCodeErr(err)
	}
	return mapper.LedgerToResp(meta), nil
}
