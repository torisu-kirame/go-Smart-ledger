package domain

import "errors"

var (
	ErrMultiNeedsTwo      = errors.New("multi ledger requires at least 2 members")
	ErrPrivateOne         = errors.New("private ledger allows only one member")
	ErrInvalidMember      = errors.New("invalid member")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrLedgerNotFound     = errors.New("ledger not found")
	ErrPendingNotFound    = errors.New("pending entry not found")
	ErrInviteNotFound     = errors.New("invite not found")
	ErrAlreadyMember      = errors.New("user is already a member")
	ErrInvalidApproval    = errors.New("invalid approval")
	ErrCannotApproveOwn   = errors.New("proposer cannot approve own entry")
)

func ValidateCreate(t LedgerType, members []Member) error {
	if len(members) == 0 {
		return ErrInvalidMember
	}
	for _, m := range members {
		if m.ID == "" {
			return ErrInvalidMember
		}
	}
	switch t {
	case LedgerPrivate:
		if len(members) != 1 {
			return ErrPrivateOne
		}
	case LedgerMulti:
		if len(members) < 2 {
			return ErrMultiNeedsTwo
		}
	default:
		return ErrInvalidMember
	}
	return nil
}

func CanAppend(meta *LedgerMeta, signerID string) error {
	for _, m := range meta.Members {
		if m.ID == signerID {
			return nil
		}
	}
	return ErrUnauthorized
}
