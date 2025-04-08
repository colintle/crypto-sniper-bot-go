package transaction

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var solanaClient = rpc.New(config.HELIUS_URL)
var wsClient *ws.Client

func init() {
	var err error
	wsClient, err = ws.Connect(context.Background(), config.HELIUS_WS)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
}

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
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
		},
	)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}

	log.Printf("🚀 Transaction submitted: %s", txHash.String())

	sub, err := wsClient.SignatureSubscribe(txHash, rpc.CommitmentFinalized)
	if err != nil {
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("Failed to subscribe to signature: %v", err)}
	}
	defer sub.Unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for {
		got, err := sub.Recv(ctx)
		if err != nil {
			log.Printf("⚠️ Subscription error: %v", err)
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Subscription error: %v", err),
			}
		}

		if got.Value.Err != nil {
			log.Printf("❌ Transaction %s failed: %v", txHash.String(), got.Value.Err)
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Transaction failed: %v", got.Value.Err),
			}
		}

		log.Printf("✅ Transaction %s confirmed!", txHash.String())
		return map[string]interface{}{
			"success": true,
			"tx_hash": txHash.String(),
		}
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
