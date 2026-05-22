package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AppendEntryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppendEntryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppendEntryLogic {
	return &AppendEntryLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppendEntryLogic) AppendEntry(id string, req *types.AppendEntryReq) (*types.EventResp, error) {
	entry := domain.EntryPayload{
		Date:     req.Entry.Date,
		Type:     req.Entry.Type,
		Amount:   req.Entry.Amount,
		Category: req.Entry.Category,
		Note:     req.Entry.Note,
	}
	ev, err := l.svcCtx.Ledger.AppendEntry(l.ctx, id, req.Entry.SignerId, entry)
	if err != nil {
		return nil, toCodeErr(err)
	}
	return &types.EventResp{
		Seq:       ev.Seq,
		Type:      ev.Type,
		Hash:      ev.Hash,
		SignerId:  ev.SignerID,
		CreatedAt: ev.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
