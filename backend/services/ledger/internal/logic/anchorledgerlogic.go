package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AnchorLedgerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnchorLedgerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnchorLedgerLogic {
	return &AnchorLedgerLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AnchorLedgerLogic) AnchorLedger(id string, req *types.AnchorReq) (*types.AnchorResp, error) {
	tx, err := l.svcCtx.Ledger.Anchor(l.ctx, id, req.SeqFrom, req.SeqTo)
	if err != nil {
		return nil, toCodeErr(err)
	}
	meta, _ := l.svcCtx.Ledger.Get(l.ctx, id)
	root := ""
	seq := req.SeqTo
	if meta != nil {
		root = meta.LatestRoot
		if seq == 0 {
			seq = meta.LatestSeq
		}
	}
	resp := &types.AnchorResp{
		LedgerId: id,
		Seq:      seq,
		Root:     root,
		TxHash:   tx,
		Status:   "synced",
	}
	if meta != nil && meta.ExternalAnchor != nil {
		resp.ExternalAnchor = externalToResp(meta.ExternalAnchor)
	}
	return resp, nil
}

func externalToResp(e *domain.ExternalAnchorRecord) *types.ExternalAnchorResp {
	if e == nil {
		return nil
	}
	return &types.ExternalAnchorResp{
		TxHash:      e.TxHash,
		ChainId:     e.ChainID,
		ChainName:   e.ChainName,
		ExplorerUrl: e.ExplorerURL,
		MerkleRoot:  e.MerkleRoot,
	}
}
