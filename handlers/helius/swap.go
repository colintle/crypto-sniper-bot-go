package helius

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"bytes"
	"strconv"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/util"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var solanaClient = rpc.New(config.SOLANA_RPC_URL)

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
	fmt.Printf("✅ Transaction successful! TxID: %s\n", txHash.String())
	return map[string]interface{}{
		"success": true,
		"tx_hash": txHash.String(),
	}
}

func Buy(tokenMint string, solSpent float64) *float64 {
	buySolAmount := solSpent / 10
	balance := util.GetSolBalance()
	if balance*config.RISK_TOLERANCE < buySolAmount {
		log.Printf("Not enough SOL. Balance: %.4f, needed: %.4f\n", balance, buySolAmount)
		return nil
	}

	lamports := int(buySolAmount * config.LAMPORTS_PER_SOL)
	quoteURL := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d", config.JUPITER_QUOTE_URL, config.SOL_MINT, tokenMint, lamports, config.SLIPPAGE)

	resp, err := http.Get(quoteURL)
	if err != nil {
		log.Printf("Quote request error: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var quoteResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		log.Printf("Quote decode error: %v", err)
		return nil
	}
	if _, ok := quoteResp["routePlan"]; !ok {
		log.Println("No route to purchase token")
		return nil
	}

	swapReq := map[string]interface{}{
		"userPublicKey": WalletPublicKey.String(),
		"quoteResponse": quoteResp,
	}
	payload, _ := json.Marshal(swapReq)
	swapResp, err := http.Post(config.JUPITER_SWAP_URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Swap request failed: %v", err)
		return nil
	}
	defer swapResp.Body.Close()
	var swapData map[string]interface{}
	json.NewDecoder(swapResp.Body).Decode(&swapData)
	if tx, ok := swapData["swapTransaction"].(string); ok {
		signAndSendTransaction(tx)
		outAmountStr, ok := quoteResp["outAmount"].(string)
		if !ok {
			log.Println("outAmount not a string")
			return nil
		}
		outAmountParsed, err := strconv.ParseFloat(outAmountStr, 64)
		if err != nil {
			log.Println("Invalid outAmount:", err)
			return nil
		}
		return &outAmountParsed
	}
	return nil
}

func Sell(tokenMint string, tokenAmount float64) map[string]interface{} {
	currentBalance := util.GetTokenBalance(tokenMint)
	if currentBalance == 0 {
		log.Println("No tokens to sell; balance is 0.")
		return nil
	}

	portionToSell := tokenAmount / 10
	var finalSellAmount float64
	if portionToSell < currentBalance {
		finalSellAmount = portionToSell
	} else {
		finalSellAmount = currentBalance
	}
	if finalSellAmount <= 0 {
		log.Println("No tokens to sell after final calculation.")
		return nil
	}

	decimals := GetTokenDecimals(tokenMint)
	if decimals == nil {
		log.Println("Missing decimals")
		return nil
	}
	lamports := int(finalSellAmount * float64(intPow(10, *decimals)))
	quoteURL := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d", config.JUPITER_QUOTE_URL, tokenMint, config.SOL_MINT, lamports, config.SLIPPAGE)

	resp, err := http.Get(quoteURL)
	if err != nil {
		log.Printf("Quote request error: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var quoteResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		log.Printf("Decode error: %v", err)
		return nil
	}
	if _, ok := quoteResp["routePlan"]; !ok {
		log.Println("No route to sell token")
		return nil
	}

	swapReq := map[string]interface{}{
		"userPublicKey": WalletPublicKey.String(),
		"quoteResponse": quoteResp,
	}
	payload, _ := json.Marshal(swapReq)
	swapResp, err := http.Post(config.JUPITER_SWAP_URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Swap request failed: %v", err)
		return nil
	}
	defer swapResp.Body.Close()
	var swapData map[string]interface{}
	json.NewDecoder(swapResp.Body).Decode(&swapData)
	if tx, ok := swapData["swapTransaction"].(string); ok {
		return signAndSendTransaction(tx)
	}
	return nil
}

func intPow(a, b int) int {
	result := 1
	for b > 0 {
		result *= a
		b--
	}
	return result
}
