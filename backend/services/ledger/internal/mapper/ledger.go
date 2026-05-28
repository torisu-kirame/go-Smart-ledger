package mapper

import (
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	"github.com/smart-ledger/go-smart-ledger/backend/services/ledger/internal/types"
)

func LedgerToResp(m *domain.LedgerMeta) *types.LedgerResp {
	members := make([]types.Member, len(m.Members))
	for i, x := range m.Members {
		members[i] = types.Member{Id: x.ID, Address: x.Address, Role: x.Role}
	}
	return &types.LedgerResp{
		Id:             m.ID,
		Type:           string(m.Type),
		Name:           m.Name,
		CreatorId:      m.CreatorID,
		LedgerAddress:  m.LedgerAddress,
		Members:         members,
		BookkeepingMode:   domain.ResolvedBookkeepingMode(m),
		MultiTableEnabled: m.MultiTableEnabled,
		Tables:            TablesToResp(m.Tables),
		EntrySchema:       EntrySchemaToResp(domain.ResolveEntrySchema(m.EntrySchema)),
		ApprovalPolicy: types.ApprovalPolicyReq{Enabled: m.ApprovalPolicy.Enabled, Threshold: m.ApprovalPolicy.Threshold},
		Encryption: types.EncryptionReq{
			Enabled:               m.Encryption.Enabled,
			Algo:                  m.Encryption.Algo,
			WrappedKeys:           m.Encryption.WrappedKeys,
			PassphraseViewEnabled: m.Encryption.PassphraseViewEnabled,
			PassphraseWrappedKeys: m.Encryption.PassphraseWrappedKeys,
		},
		LatestSeq:     m.LatestSeq,
		LatestRoot:    m.LatestRoot,
		AnchorStatus:    m.AnchorStatus,
		ExternalAnchor:  externalToResp(m.ExternalAnchor),
		LastBackupRef:     m.LastBackupRef,
		LastBackupCid:     m.LastBackupCID,
		StorageLocation: domain.NormalizeStorageLocation(m.StorageLocation),
		CreatedAt:       m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func CreateOptionsFromReq(req *types.CreateLedgerReq) ledgersvc.CreateOptions {
	return ledgersvc.CreateOptions{
		BookkeepingMode: req.BookkeepingMode,
		ApprovalPolicy: domain.ApprovalPolicy{
			Enabled:   req.ApprovalPolicy.Enabled,
			Threshold: req.ApprovalPolicy.Threshold,
		},
		Encryption: domain.LedgerEncryption{
			Enabled:     req.Encryption.Enabled,
			Algo:        req.Encryption.Algo,
			WrappedKeys: req.Encryption.WrappedKeys,
		},
		StorageLocation: req.StorageLocation,
	}
}

func externalToResp(e *domain.ExternalAnchorRecord) *types.ExternalAnchorResp {
	if e == nil {
		return nil
	}
	return &types.ExternalAnchorResp{
		TxHash:      e.TxHash,
		ChainId:     e.ChainID,
		ChainName:   e.ChainName,
		ExplorerUrl: e.ExplorerURL,
		MerkleRoot:  e.MerkleRoot,
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
