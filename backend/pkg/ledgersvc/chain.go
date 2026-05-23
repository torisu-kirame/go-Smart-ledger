package ledgersvc

import (
	"context"
	"fmt"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/txqueue"
)

// submitSteps sends transactions in order; on failure enqueues remaining steps for retry.
func (s *Service) submitSteps(ctx context.Context, label, ledgerID string, steps []miniledgerclient.TxRequest) error {
	for i, tx := range steps {
		if err := s.chain.Submit(ctx, tx); err != nil {
			if s.queue != nil {
				remaining := append([]miniledgerclient.TxRequest{}, steps[i:]...)
				_, _ = s.queue.Enqueue(label, ledgerID, remaining, err.Error())
			}
			return fmt.Errorf("miniledger submit: %w", err)
		}
	}
	return nil
}

// submitOne is a convenience for single-tx writes with queue support.
func (s *Service) submitOne(ctx context.Context, label, ledgerID string, tx miniledgerclient.TxRequest) error {
	return s.submitSteps(ctx, label, ledgerID, []miniledgerclient.TxRequest{tx})
}

// ChainQueue exposes pending chain submissions (F23).
func (s *Service) ChainQueue() []txqueue.Item {
	if s.queue == nil {
		return nil
	}
	return s.queue.List()
}

// ChainQueueStats returns pending and failed counts.
func (s *Service) ChainQueueStats() (pending, failed int) {
	if s.queue == nil {
		return 0, 0
	}
	return s.queue.Stats()
}

// RetryChainItem retries a queued submission immediately.
func (s *Service) RetryChainItem(ctx context.Context, id string) error {
	if s.queue == nil {
		return fmt.Errorf("chain queue disabled")
	}
	return s.queue.RetryNow(ctx, id)
}
