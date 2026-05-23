package ledgerhd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// Deriver derives Ethereum-style addresses (BIP44) for ledgers and members.
type Deriver struct {
	w *hdwallet.Wallet
}

func NewFromMnemonic(mnemonic string) (*Deriver, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("hd wallet mnemonic is required")
	}
	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	return &Deriver{w: w}, nil
}

// LedgerAddress is the main account for a ledger: m/44'/60'/0'/0/{ledgerIndex}
func (d *Deriver) LedgerAddress(ledgerIndex uint32) (string, error) {
	path, err := hdwallet.ParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", ledgerIndex))
	if err != nil {
		return "", err
	}
	acct, err := d.w.Derive(path, false)
	if err != nil {
		return "", err
	}
	return acct.Address.Hex(), nil
}

// MemberAddress derives per-member address: m/44'/60'/0'/1/{ledgerIndex}/{memberIndex}
func (d *Deriver) MemberAddress(ledgerIndex, memberIndex uint32) (string, error) {
	path, err := hdwallet.ParseDerivationPath(
		fmt.Sprintf("m/44'/60'/0'/1/%d/%d", ledgerIndex, memberIndex),
	)
	if err != nil {
		return "", err
	}
	acct, err := d.w.Derive(path, false)
	if err != nil {
		return "", err
	}
	return acct.Address.Hex(), nil
}

// NormalizeAddress lowercases hex checksummed address.
func NormalizeAddress(addr string) string {
	if common.IsHexAddress(addr) {
		return common.HexToAddress(addr).Hex()
	}
	return addr
}
