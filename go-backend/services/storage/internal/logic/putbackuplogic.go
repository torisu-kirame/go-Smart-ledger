package logic

import (
	"context"
	"errors"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/storage"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type PutBackupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPutBackupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutBackupLogic {
	return &PutBackupLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *PutBackupLogic) PutBackup(req *types.PutBackupReq) (*types.PutBackupResp, error) {
	if req.LedgerId == "" || req.Password == "" {
		return nil, xerrors.New(400, "ledgerId and password required")
	}
	plain, err := storage.DecodeB64(req.Payload)
	if err != nil {
		return nil, xerrors.New(400, "invalid base64 payload")
	}
	ref, err := l.svcCtx.Backup.Put(l.ctx, req.LedgerId, req.Password, plain)
	if err != nil {
		return nil, xerrors.New(500, err.Error())
	}
	return &types.PutBackupResp{Ref: ref}, nil
}

func backupErr(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return xerrors.New(404, err.Error())
	}
	return xerrors.New(500, err.Error())
}
