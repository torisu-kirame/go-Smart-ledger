package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListEventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEventsLogic {
	return &ListEventsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListEventsLogic) ListEvents(id string, req *types.ListEventsReq) ([]types.EventResp, error) {
	from := req.From
	if from == 0 {
		from = 1
	}
	events, err := l.svcCtx.Ledger.ListEvents(l.ctx, id, from, req.To)
	if err != nil {
		return nil, toCodeErr(err)
	}
	out := make([]types.EventResp, len(events))
	for i, e := range events {
		out[i] = types.EventResp{
			Seq:       e.Seq,
			Type:      e.Type,
			Hash:      e.Hash,
			SignerId:  e.SignerID,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return out, nil
}
