package chainstore

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	fiscotypes "github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

func fiscoNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func signFiscoContractCall(
	ctx context.Context,
	http *fiscoHTTPClient,
	cfg FISCOConfig,
	contract common.Address,
	input []byte,
	key *ecdsa.PrivateKey,
) (string, error) {
	height, err := http.getBlockNumber(ctx)
	if err != nil {
		return "", err
	}
	to := contract
	tx := &fiscotypes.Transaction{
		SMCrypto: cfg.IsSMCrypto,
		Data: fiscotypes.TransactionData{
			Version:    0,
			ChainID:    cfg.ChainID,
			GroupID:    cfg.GroupID,
			BlockLimit: int64(height) + fiscoBlockLimitDelta,
			Nonce:      fiscoNonce(),
			To:         &to,
			Input:      append([]byte(nil), input...),
		},
	}
	signer := fiscotypes.FrontierSigner{}
	if cfg.IsSMCrypto {
		return "", fmt.Errorf("fisco: IsSMCrypto not supported in pure-Go client yet")
	}
	signed, err := fiscotypes.SignTx(tx, signer, key)
	if err != nil {
		return "", fmt.Errorf("fisco: sign tx: %w", err)
	}
	return "0x" + hex.EncodeToString(signed.Bytes()), nil
}
