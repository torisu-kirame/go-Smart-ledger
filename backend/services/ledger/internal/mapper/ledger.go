package mapper

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
)

func LedgerToResp(m *domain.LedgerMeta) *types.LedgerResp {
	members := make([]types.Member, len(m.Members))
	for i, x := range m.Members {
		members[i] = types.Member{Id: x.ID, Address: x.Address, Role: x.Role}
	}
	return &types.LedgerResp{
		Id:           m.ID,
		Type:         string(m.Type),
		Name:         m.Name,
		CreatorId:    m.CreatorID,
		Members:      members,
		LatestSeq:    m.LatestSeq,
		LatestRoot:   m.LatestRoot,
		AnchorStatus: m.AnchorStatus,
	}
}

func ParseLedgerType(s string) (domain.LedgerType, error) {
	switch s {
	case "private":
		return domain.LedgerPrivate, nil
	case "multi":
		return domain.LedgerMulti, nil
	default:
		return "", domain.ErrInvalidMember
	}
}

func MembersFromReq(in []types.Member) []domain.Member {
	out := make([]domain.Member, len(in))
	for i, m := range in {
		out[i] = domain.Member{ID: m.Id, Address: m.Address, Role: m.Role}
	}
	return out
}
