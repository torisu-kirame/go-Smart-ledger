package logic

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/storage"
	"github.com/smart-ledger/go-smart-ledger/backend/services/storage/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/storage/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type GetBackupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBackupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBackupLogic {
	return &GetBackupLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetBackupLogic) GetBackup(req *types.GetBackupReq) (*types.GetBackupResp, error) {
	if req.Ref == "" || req.Password == "" {
		return nil, xerrors.New(400, "ref and password required")
	}
	if err := storage.ValidateRef(req.Ref); err != nil {
		return nil, xerrors.New(400, err.Error())
	}
	plain, err := l.svcCtx.Backup.Get(l.ctx, req.Ref, req.Password)
	if err != nil {
		return nil, backupErr(err)
	}
	return &types.GetBackupResp{Payload: storage.EncodeB64(plain)}, nil
}
