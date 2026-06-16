package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/services/gateway/internal/svc"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (*types.HealthResp, error) {
	resp := &types.HealthResp{
		Status:           "ok",
		Gateway:          "ok",
		MiniLedgerOnline: false,
	}
	ledgerURL := strings.TrimRight(l.svcCtx.Config.Upstreams.Ledger, "/") + "/api/v1/health"
	req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, ledgerURL, nil)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("health: build ledger request: %v", err)
		return resp, nil
	}
	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logx.WithContext(l.ctx).Infof("health: ledger unreachable: %v", err)
		return resp, nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logx.WithContext(l.ctx).Infof("health: ledger status %d", res.StatusCode)
		return resp, nil
	}
	var ledgerHealth struct {
		MiniLedgerOnline bool `json:"miniLedgerOnline"`
		QueuePending     int  `json:"queuePending"`
		QueueFailed      int  `json:"queueFailed"`
	}
	if err := json.NewDecoder(res.Body).Decode(&ledgerHealth); err != nil {
		logx.WithContext(l.ctx).Errorf("health: decode ledger response: %v", err)
		return resp, nil
	}
	resp.MiniLedgerOnline = ledgerHealth.MiniLedgerOnline
	resp.ChainQueuePending = ledgerHealth.QueuePending
	resp.ChainQueueFailed = ledgerHealth.QueueFailed
	return resp, nil
}
