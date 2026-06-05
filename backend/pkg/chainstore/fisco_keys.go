package chainstore

import (
	"fmt"
	"strings"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
)

const fiscoGlobalLedgerID = "__global__"

func fiscoGlobalLedgerIDsKey() string { return "smartledger:__ledger_ids__" }

func fiscoLedgerKeysIndexKey(ledgerID string) string {
	return fmt.Sprintf("smartledger:ledger:%s:__keys__", ledgerID)
}

func fiscoInviteIndexKey(inviteeID string) string {
	return fmt.Sprintf("smartledger:__index__:invite:%s", inviteeID)
}

func isInternalFiscoKey(key string) bool {
	return strings.HasSuffix(key, ":__keys__") ||
		strings.HasPrefix(key, "smartledger:__index__:") ||
		key == fiscoGlobalLedgerIDsKey()
}

// ledgerIDFromStateKey maps a world_state key to the contract ledgerId partition.
func ledgerIDFromStateKey(key string) string {
	const p = "smartledger:ledger:"
	if !strings.HasPrefix(key, p) {
		return fiscoGlobalLedgerID
	}
	rest := key[len(p):]
	if i := strings.IndexByte(rest, ':'); i >= 0 {
		return rest[:i]
	}
	return rest
}

func isLedgerMetaKey(key string) bool {
	p := "smartledger:ledger:"
	if !strings.HasPrefix(key, p) {
		return false
	}
	rest := key[len(p):]
	return !strings.Contains(rest, ":")
}

func inviteeFromInviteKey(key string) string {
	const marker = ":invite:"
	if i := strings.LastIndex(key, marker); i >= 0 {
		return key[i+len(marker):]
	}
	return ""
}

func ledgerMetaKeyForID(id string) string { return domain.LedgerMetaKey(id) }
