package domain

import "time"

// ApprovalPolicy configures multi-member entry approval (F17).
type ApprovalPolicy struct {
	Enabled   bool `json:"enabled"`
	Threshold int  `json:"threshold"` // approvals required before EntryAdded
}

// LedgerEncryption holds per-member wrapped group keys (F19, client-side crypto).
type LedgerEncryption struct {
	Enabled     bool              `json:"enabled"`
	Algo        string            `json:"algo,omitempty"` // aes-gcm-v1
	WrappedKeys map[string]string `json:"wrappedKeys,omitempty"`
}

// PendingEntry is stored off-chain in world_state until approved or rejected.
type PendingEntry struct {
	ID         string    `json:"id"`
	LedgerID   string    `json:"ledgerId"`
	ProposerID string    `json:"proposerId"`
	Payload    []byte    `json:"payload"`
	Approvals  []string  `json:"approvals"`
	Status     string    `json:"status"` // pending | approved | rejected
	CreatedAt  time.Time `json:"createdAt"`
}

// MemberInvite invites a user to join a ledger (F18).
type MemberInvite struct {
	LedgerID   string    `json:"ledgerId"`
	InviteeID  string    `json:"inviteeId"`
	InviterID  string    `json:"inviterId"`
	Role       string    `json:"role,omitempty"`
	Status     string    `json:"status"` // pending | accepted | revoked
	CreatedAt  time.Time `json:"createdAt"`
}

const (
	PendingStatusPending  = "pending"
	PendingStatusApproved = "approved"
	PendingStatusRejected = "rejected"

	InviteStatusPending = "pending"

	EventEntryProposed   = "EntryProposed"
	EventEntryApproved   = "EntryApproved"
	EventEntryRejected   = "EntryRejected"
	EventMemberInvited   = "MemberInvited"
	EventMemberJoined    = "MemberJoined"
	EventGroupKeyRotated = "GroupKeyRotated"
)

func DefaultApprovalPolicy(t LedgerType, memberCount int) ApprovalPolicy {
	if t != LedgerMulti || memberCount < 2 {
		return ApprovalPolicy{Enabled: false, Threshold: 1}
	}
	th := 2
	if memberCount < th {
		th = memberCount
	}
	return ApprovalPolicy{Enabled: true, Threshold: th}
}

func IsMember(meta *LedgerMeta, userID string) bool {
	if meta == nil || userID == "" {
		return false
	}
	for _, m := range meta.Members {
		if m.ID == userID {
			return true
		}
	}
	return false
}

func ApprovalRequired(meta *LedgerMeta) bool {
	if meta == nil || meta.Type != LedgerMulti {
		return false
	}
	if meta.ApprovalPolicy.Enabled {
		return meta.ApprovalPolicy.Threshold > 1
	}
	return false
}

func HasApproved(p *PendingEntry, userID string) bool {
	for _, id := range p.Approvals {
		if id == userID {
			return true
		}
	}
	return false
}
