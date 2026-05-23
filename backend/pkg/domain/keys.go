package domain

import "fmt"

const keyPrefix = "smartledger"

func LedgerMetaKey(id string) string {
	return fmt.Sprintf("%s:ledger:%s", keyPrefix, id)
}

func LedgerEventKey(id string, seq uint64) string {
	return fmt.Sprintf("%s:ledger:%s:event:%d", keyPrefix, id, seq)
}

func LedgerIndexPrefix() string {
	return keyPrefix + ":ledger:%"
}

func LedgerPendingKey(ledgerID, pendingID string) string {
	return fmt.Sprintf("%s:ledger:%s:pending:%s", keyPrefix, ledgerID, pendingID)
}

func LedgerPendingPrefix(ledgerID string) string {
	return fmt.Sprintf("%s:ledger:%s:pending:", keyPrefix, ledgerID)
}

func LedgerInviteKey(ledgerID, inviteeID string) string {
	return fmt.Sprintf("%s:ledger:%s:invite:%s", keyPrefix, ledgerID, inviteeID)
}

func LedgerInviteSuffix(inviteeID string) string {
	return ":invite:" + inviteeID
}
