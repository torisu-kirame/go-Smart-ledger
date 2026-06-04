package chainstore

import (
	"context"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/miniledgerclient"
)

type miniLedgerStore struct {
	client *miniledgerclient.Client
}

func newMiniLedger(baseURL string) Store {
	return &miniLedgerStore{client: miniledgerclient.New(baseURL)}
}

func (a *miniLedgerStore) Backend() Backend { return BackendMiniLedger }

func (a *miniLedgerStore) Ping(ctx context.Context) error { return a.client.Ping(ctx) }

func (a *miniLedgerStore) Status(ctx context.Context) (*Status, error) {
	st, err := a.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &Status{
		Height:  st.Height,
		Uptime:  st.Uptime,
		Role:    st.Role,
		Backend: string(BackendMiniLedger),
	}, nil
}

func (a *miniLedgerStore) Submit(ctx context.Context, tx TxRequest) error {
	return a.client.Submit(ctx, miniledgerclient.TxRequest{
		Key: tx.Key, Value: tx.Value, Type: tx.Type, Payload: tx.Payload,
	})
}

func (a *miniLedgerStore) Query(ctx context.Context, sql string, params ...interface{}) ([]StateRow, error) {
	rows, err := a.client.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	out := make([]StateRow, len(rows))
	for i, r := range rows {
		out[i] = StateRow{Key: r.Key, Value: r.Value}
	}
	return out, nil
}

func (a *miniLedgerStore) GetRaw(ctx context.Context, path string) ([]byte, error) {
	return a.client.GetRaw(ctx, path)
}

func (a *miniLedgerStore) BaseURL() string { return a.client.BaseURL() }
