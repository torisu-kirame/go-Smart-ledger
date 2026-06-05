package chainstore

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type fiscoParsedQuery struct {
	eqKey       string
	likePrefix  string
	notLike     string
	suffixMatch string // key LIKE %:invite:user
}

func parseFiscoQuery(sql string, params []interface{}) (fiscoParsedQuery, error) {
	norm := strings.Join(strings.Fields(sql), " ")
	var q fiscoParsedQuery
	switch {
	case strings.Contains(norm, "key = ?"):
		if len(params) < 1 {
			return q, fmt.Errorf("fisco query: missing key param")
		}
		q.eqKey, _ = params[0].(string)
		if q.eqKey == "" {
			return q, fmt.Errorf("fisco query: empty key")
		}
	case strings.Contains(norm, "key NOT LIKE ?"):
		if len(params) < 2 {
			return q, fmt.Errorf("fisco query: LIKE+NOT LIKE needs 2 params")
		}
		q.likePrefix, _ = params[0].(string)
		q.notLike, _ = params[1].(string)
	case strings.Contains(norm, "key LIKE ?"):
		if len(params) < 1 {
			return q, fmt.Errorf("fisco query: missing LIKE param")
		}
		pat, _ := params[0].(string)
		if strings.HasPrefix(pat, "%") {
			q.suffixMatch = pat
		} else {
			q.likePrefix = pat
		}
	default:
		return q, fmt.Errorf("fisco query: unsupported SQL: %s", sql)
	}
	return q, nil
}

func (s *FISCOStore) runQuery(ctx context.Context, q fiscoParsedQuery) ([]StateRow, error) {
	if s.registry == nil {
		return nil, fmt.Errorf("fisco: RegistryContract not configured")
	}
	switch {
	case q.eqKey != "":
		return s.queryEq(ctx, q.eqKey)
	case q.suffixMatch != "":
		return s.querySuffix(ctx, q.suffixMatch)
	case q.likePrefix != "" && q.notLike != "":
		return s.queryListLedgers(ctx, q.likePrefix, q.notLike)
	case q.likePrefix != "":
		return s.queryLikePrefix(ctx, q.likePrefix)
	default:
		return nil, fmt.Errorf("fisco query: empty pattern")
	}
}

func (s *FISCOStore) queryEq(ctx context.Context, key string) ([]StateRow, error) {
	ledgerID := ledgerIDFromStateKey(key)
	raw, err := s.registry.getState(ctx, ledgerID, key)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	return []StateRow{{Key: key, Value: json.RawMessage(raw)}}, nil
}

func (s *FISCOStore) queryLikePrefix(ctx context.Context, prefix string) ([]StateRow, error) {
	ledgerID := ledgerIDFromLikePrefix(prefix)
	if ledgerID == "" {
		return nil, fmt.Errorf("fisco query: cannot resolve ledger from prefix %q", prefix)
	}
	idxKey := fiscoLedgerKeysIndexKey(ledgerID)
	keys, err := s.registry.getJSONIndex(ctx, ledgerID, idxKey)
	if err != nil {
		return nil, err
	}
	var out []StateRow
	for _, k := range keys {
		if isInternalFiscoKey(k) {
			continue
		}
		if !strings.HasPrefix(k, strings.TrimSuffix(prefix, "%")) {
			continue
		}
		row, err := s.fetchRow(ctx, k)
		if err != nil || row == nil {
			continue
		}
		out = append(out, *row)
	}
	slices.SortFunc(out, func(a, b StateRow) int { return strings.Compare(a.Key, b.Key) })
	return out, nil
}

func (s *FISCOStore) queryListLedgers(ctx context.Context, likePrefix, notLike string) ([]StateRow, error) {
	ids, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, fiscoGlobalLedgerIDsKey())
	if err != nil {
		return nil, err
	}
	var out []StateRow
	eventExclude := strings.TrimSuffix(notLike, "%")
	for _, id := range ids {
		metaKey := ledgerMetaKeyForID(id)
		if !strings.HasPrefix(metaKey, strings.TrimSuffix(likePrefix, "%")) {
			continue
		}
		if strings.Contains(metaKey, eventExclude) {
			continue
		}
		row, err := s.fetchRow(ctx, metaKey)
		if err != nil || row == nil {
			continue
		}
		out = append(out, *row)
	}
	slices.SortFunc(out, func(a, b StateRow) int { return strings.Compare(a.Key, b.Key) })
	return out, nil
}

func (s *FISCOStore) querySuffix(ctx context.Context, pattern string) ([]StateRow, error) {
	suffix := strings.TrimPrefix(pattern, "%")
	if !strings.Contains(suffix, ":invite:") {
		return nil, fmt.Errorf("fisco query: unsupported suffix pattern %q", pattern)
	}
	invitee := strings.TrimPrefix(suffix, ":invite:")
	idxKey := fiscoInviteIndexKey(invitee)
	keys, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, idxKey)
	if err != nil {
		return nil, err
	}
	var out []StateRow
	for _, k := range keys {
		if !strings.HasSuffix(k, suffix) {
			continue
		}
		row, err := s.fetchRow(ctx, k)
		if err != nil || row == nil {
			continue
		}
		out = append(out, *row)
	}
	slices.SortFunc(out, func(a, b StateRow) int { return strings.Compare(a.Key, b.Key) })
	return out, nil
}

func (s *FISCOStore) fetchRow(ctx context.Context, key string) (*StateRow, error) {
	ledgerID := ledgerIDFromStateKey(key)
	raw, err := s.registry.getState(ctx, ledgerID, key)
	if err != nil || len(raw) == 0 {
		return nil, err
	}
	return &StateRow{Key: key, Value: json.RawMessage(raw)}, nil
}

func ledgerIDFromLikePrefix(prefix string) string {
	p := strings.TrimSuffix(prefix, "%")
	const marker = "smartledger:ledger:"
	if !strings.HasPrefix(p, marker) {
		return ""
	}
	rest := p[len(marker):]
	if i := strings.IndexByte(rest, ':'); i >= 0 {
		return rest[:i]
	}
	return rest
}
