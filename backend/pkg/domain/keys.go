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

func LedgerCOAKey(ledgerID string) string {
	return fmt.Sprintf("%s:ledger:%s:coa", keyPrefix, ledgerID)
}

func LedgerPeriodKey(ledgerID, period string) string {
	return fmt.Sprintf("%s:ledger:%s:period:%s", keyPrefix, ledgerID, period)
}

func LedgerPeriodPrefix(ledgerID string) string {
	return fmt.Sprintf("%s:ledger:%s:period:", keyPrefix, ledgerID)
}

func LedgerBankStmtKey(ledgerID, stmtID string) string {
	return fmt.Sprintf("%s:ledger:%s:bank:%s", keyPrefix, ledgerID, stmtID)
}

func LedgerBankStmtPrefix(ledgerID string) string {
	return fmt.Sprintf("%s:ledger:%s:bank:", keyPrefix, ledgerID)
}

func LedgerAttachmentKey(ledgerID string, entrySeq uint64, attachID string) string {
	return fmt.Sprintf("%s:ledger:%s:attach:%d:%s", keyPrefix, ledgerID, entrySeq, attachID)
}

func LedgerAttachmentPrefix(ledgerID string) string {
	return fmt.Sprintf("%s:ledger:%s:attach:", keyPrefix, ledgerID)
}
