// Package txqueue persists failed MiniLedger submissions and retries them in the background.
package txqueue

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/snowflake"
)

const (
	StatusPending  = "pending"
	StatusRetrying = "retrying"
	StatusFailed   = "failed"
	StatusDone     = "done"
)

// Item is a multi-step chain submission that failed mid-flight.
type Item struct {
	ID         string                      `json:"id"`
	Label      string                      `json:"label"`
	LedgerID   string                      `json:"ledgerId,omitempty"`
	Steps      []miniledgerclient.TxRequest `json:"steps"`
	Status     string                      `json:"status"`
	Attempts   int                         `json:"attempts"`
	LastError  string                      `json:"lastError,omitempty"`
	CreatedAt  time.Time                   `json:"createdAt"`
	UpdatedAt  time.Time                   `json:"updatedAt"`
}

// SubmitFunc submits a single transaction to MiniLedger.
type SubmitFunc func(ctx context.Context, tx miniledgerclient.TxRequest) error

// Queue stores pending chain writes and retries them periodically.
type Queue struct {
	mu          sync.Mutex
	items       map[string]*Item
	path        string
	submit      SubmitFunc
	maxAttempts int
	interval    time.Duration
}

// Options configures the retry queue.
type Options struct {
	PersistPath string
	MaxAttempts int
	Interval    time.Duration
}

func New(submit SubmitFunc, opt Options) (*Queue, error) {
	if submit == nil {
		return nil, errors.New("txqueue: submit func required")
	}
	if opt.MaxAttempts <= 0 {
		opt.MaxAttempts = 30
	}
	if opt.Interval <= 0 {
		opt.Interval = 5 * time.Second
	}
	q := &Queue{
		items:       make(map[string]*Item),
		path:        opt.PersistPath,
		submit:      submit,
		maxAttempts: opt.MaxAttempts,
		interval:    opt.Interval,
	}
	if opt.PersistPath != "" {
		if err := os.MkdirAll(filepath.Dir(opt.PersistPath), 0o755); err != nil {
			return nil, err
		}
		_ = q.load()
	}
	return q, nil
}

// Enqueue adds remaining steps after a partial failure.
func (q *Queue) Enqueue(label, ledgerID string, steps []miniledgerclient.TxRequest, cause string) (*Item, error) {
	if len(steps) == 0 {
		return nil, errors.New("txqueue: empty steps")
	}
	now := time.Now().UTC()
	id, err := snowflake.NextString()
	if err != nil {
		return nil, err
	}
	it := &Item{
		ID:        id,
		Label:     label,
		LedgerID:  ledgerID,
		Steps:     steps,
		Status:    StatusPending,
		LastError: cause,
		CreatedAt: now,
		UpdatedAt: now,
	}
	q.mu.Lock()
	q.items[it.ID] = it
	q.mu.Unlock()
	if err := q.save(); err != nil {
		return it, err
	}
	return it, nil
}

// List returns a snapshot of non-done items (newest first).
func (q *Queue) List() []Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Item, 0, len(q.items))
	for _, it := range q.items {
		if it.Status == StatusDone {
			continue
		}
		out = append(out, *it)
	}
	// simple sort by UpdatedAt desc
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].UpdatedAt.After(out[i].UpdatedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// Stats returns counts by status.
func (q *Queue) Stats() (pending, failed int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, it := range q.items {
		switch it.Status {
		case StatusPending, StatusRetrying:
			pending++
		case StatusFailed:
			failed++
		}
	}
	return pending, failed
}

// RetryNow forces one immediate attempt for an item.
func (q *Queue) RetryNow(ctx context.Context, id string) error {
	q.mu.Lock()
	it, ok := q.items[id]
	if !ok {
		q.mu.Unlock()
		return errors.New("txqueue: item not found")
	}
	if it.Status == StatusDone {
		q.mu.Unlock()
		return nil
	}
	it.Status = StatusRetrying
	q.mu.Unlock()
	return q.flushOne(ctx, it)
}

// StartWorker runs until ctx is cancelled.
func (q *Queue) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(q.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.flushAll(ctx)
		}
	}
}

func (q *Queue) flushAll(ctx context.Context) {
	q.mu.Lock()
	ids := make([]string, 0, len(q.items))
	for id, it := range q.items {
		if it.Status == StatusPending || it.Status == StatusRetrying {
			ids = append(ids, id)
		}
	}
	q.mu.Unlock()
	for _, id := range ids {
		q.mu.Lock()
		it, ok := q.items[id]
		q.mu.Unlock()
		if !ok {
			continue
		}
		_ = q.flushOne(ctx, it)
	}
}

func (q *Queue) flushOne(ctx context.Context, it *Item) error {
	q.mu.Lock()
	it.Status = StatusRetrying
	it.Attempts++
	it.UpdatedAt = time.Now().UTC()
	q.mu.Unlock()

	for len(it.Steps) > 0 {
		if err := q.submit(ctx, it.Steps[0]); err != nil {
			q.mu.Lock()
			it.LastError = err.Error()
			it.UpdatedAt = time.Now().UTC()
			if it.Attempts >= q.maxAttempts {
				it.Status = StatusFailed
			} else {
				it.Status = StatusPending
			}
			q.mu.Unlock()
			_ = q.save()
			return err
		}
		it.Steps = it.Steps[1:]
	}

	q.mu.Lock()
	it.Status = StatusDone
	it.Steps = nil
	it.UpdatedAt = time.Now().UTC()
	delete(q.items, it.ID)
	q.mu.Unlock()
	return q.save()
}

func (q *Queue) load() error {
	raw, err := os.ReadFile(q.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var items []Item
	if err := json.Unmarshal(raw, &items); err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range items {
		it := items[i]
		if it.Status != StatusDone {
			cp := it
			q.items[it.ID] = &cp
		}
	}
	return nil
}

func (q *Queue) save() error {
	if q.path == "" {
		return nil
	}
	q.mu.Lock()
	items := make([]Item, 0, len(q.items))
	for _, it := range q.items {
		items = append(items, *it)
	}
	q.mu.Unlock()
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	tmp := q.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, q.path)
}
