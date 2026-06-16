package chainstore

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
)

func (s *FISCOStore) syncIndexes(ctx context.Context, stateKey string, value []byte, delete bool) error {
	if s.registry == nil {
		return nil
	}
	ledgerID := ledgerIDFromStateKey(stateKey)

	if delete {
		if err := s.removeFromLedgerKeyIndex(ctx, ledgerID, stateKey); err != nil {
			return err
		}
		if invitee := inviteeFromInviteKey(stateKey); invitee != "" {
			if err := s.removeFromInviteIndex(ctx, invitee, stateKey); err != nil {
				return err
			}
		}
		if isLedgerMetaKey(stateKey) {
			return s.removeGlobalLedgerID(ctx, ledgerIDFromStateKey(stateKey))
		}
		return nil
	}

	if err := s.addToLedgerKeyIndex(ctx, ledgerID, stateKey); err != nil {
		return err
	}
	if isLedgerMetaKey(stateKey) {
		id := ledgerIDFromStateKey(stateKey)
		if err := s.addGlobalLedgerID(ctx, id); err != nil {
			return err
		}
	}
	if invitee := inviteeFromInviteKey(stateKey); invitee != "" {
		if err := s.addToInviteIndex(ctx, invitee, stateKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *FISCOStore) addToLedgerKeyIndex(ctx context.Context, ledgerID, stateKey string) error {
	idxKey := fiscoLedgerKeysIndexKey(ledgerID)
	keys, err := s.registry.getJSONIndex(ctx, ledgerID, idxKey)
	if err != nil {
		return err
	}
	if slices.Contains(keys, stateKey) {
		return nil
	}
	keys = append(keys, stateKey)
	return s.registry.putJSONIndex(ctx, ledgerID, idxKey, keys)
}

func (s *FISCOStore) removeFromLedgerKeyIndex(ctx context.Context, ledgerID, stateKey string) error {
	idxKey := fiscoLedgerKeysIndexKey(ledgerID)
	keys, err := s.registry.getJSONIndex(ctx, ledgerID, idxKey)
	if err != nil || len(keys) == 0 {
		return err
	}
	keys = slices.DeleteFunc(keys, func(k string) bool { return k == stateKey })
	return s.registry.putJSONIndex(ctx, ledgerID, idxKey, keys)
}

func (s *FISCOStore) addGlobalLedgerID(ctx context.Context, id string) error {
	keys, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, fiscoGlobalLedgerIDsKey())
	if err != nil {
		return err
	}
	if slices.Contains(keys, id) {
		return nil
	}
	keys = append(keys, id)
	return s.registry.putJSONIndex(ctx, fiscoGlobalLedgerID, fiscoGlobalLedgerIDsKey(), keys)
}

func (s *FISCOStore) removeGlobalLedgerID(ctx context.Context, id string) error {
	keys, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, fiscoGlobalLedgerIDsKey())
	if err != nil || len(keys) == 0 {
		return err
	}
	keys = slices.DeleteFunc(keys, func(k string) bool { return k == id })
	return s.registry.putJSONIndex(ctx, fiscoGlobalLedgerID, fiscoGlobalLedgerIDsKey(), keys)
}

func (s *FISCOStore) addToInviteIndex(ctx context.Context, inviteeID, stateKey string) error {
	idxKey := fiscoInviteIndexKey(inviteeID)
	keys, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, idxKey)
	if err != nil {
		return err
	}
	if slices.Contains(keys, stateKey) {
		return nil
	}
	keys = append(keys, stateKey)
	return s.registry.putJSONIndex(ctx, fiscoGlobalLedgerID, idxKey, keys)
}

func (s *FISCOStore) removeFromInviteIndex(ctx context.Context, inviteeID, stateKey string) error {
	idxKey := fiscoInviteIndexKey(inviteeID)
	keys, err := s.registry.getJSONIndex(ctx, fiscoGlobalLedgerID, idxKey)
	if err != nil || len(keys) == 0 {
		return err
	}
	keys = slices.DeleteFunc(keys, func(k string) bool { return k == stateKey })
	return s.registry.putJSONIndex(ctx, fiscoGlobalLedgerID, idxKey, keys)
}

func txValueIsDelete(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	s := strings.TrimSpace(string(raw))
	return s == "null" || s == ""
}
