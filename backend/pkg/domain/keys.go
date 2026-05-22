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
