package chainstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

const migrateKeyPrefix = "smartledger:%"

// MigrateOptions controls MiniLedger → FISCO KV migration.
type MigrateOptions struct {
	DryRun   bool
	Verify   bool
	Limit    int
	OnProgress func(done, total int, key string)
}

// MigrateResult summarizes a migration run.
type MigrateResult struct {
	Total   int `json:"total"`
	Written int `json:"written"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
	Verified int `json:"verified,omitempty"`
	Mismatch int `json:"mismatch,omitempty"`
}

// MigrateMiniLedgerToFISCO copies world_state keys from src to dst via dst.Submit.
// Index blobs on FISCO are rebuilt automatically by FISCOStore.Submit.
func MigrateMiniLedgerToFISCO(ctx context.Context, src, dst Store, opt MigrateOptions) (*MigrateResult, error) {
	if src == nil {
		return nil, fmt.Errorf("migrate: src store required")
	}
	if !opt.DryRun {
		if dst == nil {
			return nil, fmt.Errorf("migrate: dst store required")
		}
		if dst.Backend() != BackendFISCO {
			return nil, fmt.Errorf("migrate: dst must be fisco backend, got %q", dst.Backend())
		}
	}
	rows, err := src.Query(ctx,
		`SELECT key, value FROM world_state WHERE key LIKE ? ORDER BY key`,
		migrateKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("migrate: export miniledger: %w", err)
	}
	rows = filterMigrateRows(rows)
	sortMigrateRows(rows)
	if opt.Limit > 0 && len(rows) > opt.Limit {
		rows = rows[:opt.Limit]
	}

	res := &MigrateResult{Total: len(rows)}
	for i, row := range rows {
		if opt.OnProgress != nil {
			opt.OnProgress(i+1, len(rows), row.Key)
		}
		if opt.DryRun {
			res.Written++
			continue
		}
		if err := dst.Submit(ctx, TxRequest{Key: row.Key, Value: row.Value}); err != nil {
			res.Failed++
			return res, fmt.Errorf("migrate: key %q: %w", row.Key, err)
		}
		res.Written++
	}

	if opt.Verify && !opt.DryRun {
		if dst == nil {
			return res, fmt.Errorf("migrate: dst required for verify")
		}
		verified, mismatch, err := verifyMigratedRows(ctx, rows, dst)
		res.Verified = verified
		res.Mismatch = mismatch
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func filterMigrateRows(rows []StateRow) []StateRow {
	out := make([]StateRow, 0, len(rows))
	for _, row := range rows {
		if row.Key == "" {
			continue
		}
		if shouldSkipMigrateKey(row.Key) {
			continue
		}
		out = append(out, row)
	}
	return out
}

func shouldSkipMigrateKey(key string) bool {
	if isInternalFiscoKey(key) {
		return true
	}
	return strings.Contains(key, ":__keys__")
}

func sortMigrateRows(rows []StateRow) {
	slices.SortFunc(rows, func(a, b StateRow) int {
		ma, mb := isLedgerMetaKey(a.Key), isLedgerMetaKey(b.Key)
		if ma != mb {
			if ma {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Key, b.Key)
	})
}

func verifyMigratedRows(ctx context.Context, rows []StateRow, dst Store) (verified, mismatch int, err error) {
	for _, row := range rows {
		if txValueIsDelete(row.Value) {
			continue
		}
		got, err := dst.Query(ctx, `SELECT key, value FROM world_state WHERE key = ?`, row.Key)
		if err != nil {
			return verified, mismatch, fmt.Errorf("migrate verify %q: %w", row.Key, err)
		}
		if len(got) == 0 {
			mismatch++
			continue
		}
		if !stateValuesEqual(row.Value, got[0].Value) {
			mismatch++
			continue
		}
		verified++
	}
	if mismatch > 0 {
		return verified, mismatch, fmt.Errorf("migrate: %d key(s) mismatch after write", mismatch)
	}
	return verified, mismatch, nil
}

func stateValuesEqual(a, b json.RawMessage) bool {
	if bytes.Equal(bytes.TrimSpace(a), bytes.TrimSpace(b)) {
		return true
	}
	var va, vb any
	if json.Unmarshal(a, &va) != nil || json.Unmarshal(b, &vb) != nil {
		return false
	}
	ra, _ := json.Marshal(va)
	rb, _ := json.Marshal(vb)
	return bytes.Equal(ra, rb)
}
