package transaction

import (
	"context"

	"github.com/colintle/crypto-sniper-bot-go/config"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var solanaClient = rpc.New(config.HELIUS_URL)

func signAndSendTransaction(swapTxBase64 string) map[string]interface{} {
	tx, err := solana.TransactionFromBase64(swapTxBase64)
	if err != nil {
		return map[string]interface{}{"success": false, "error": "Invalid transaction"}
	}
	tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(WalletPublicKey) {
			return &WalletPrivateKey
		}
		return nil
	})
	txHash, err := solanaClient.SendTransactionWithOpts(
		context.Background(),
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       true,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{
		"success": true,
		"tx_hash": txHash.String(),
	}
}

func intPow(a, b int) int {
	result := 1
	for b > 0 {
		result *= a
		b--
	}
	return result
}
