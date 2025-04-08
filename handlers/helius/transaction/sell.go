package transaction

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/database"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/operations"
	"github.com/colintle/crypto-sniper-bot-go/models"
)

func Sell(tokenMint string, tokenAmount float64, decimals int) models.Trade {
	currentBalance := operations.GetTokenBalanceCached(tokenMint)
	currentSOLBalance := operations.GetSOLBalanceCached()
	if currentBalance == 0 {
		return tradeError("No tokens to sell; balance is 0.", tokenMint, models.Sell, models.LOW_TOKEN, currentSOLBalance)
	}

	portionToSell := tokenAmount / config.POSITION_SIZE
	finalSellAmount := min(portionToSell, currentBalance)

	if math.Abs(currentBalance-portionToSell) < 5 {
		finalSellAmount = currentBalance
	}

	if finalSellAmount <= 0 {
		return tradeError("No tokens to sell after final calculation.", tokenMint, models.Sell, models.MIN_TOKEN, currentSOLBalance)
	}

	lamports := int(finalSellAmount * float64(intPow(10, decimals)))

	quoteChan := make(chan map[string]interface{})
	doneChan := make(chan struct{})
	defer close(doneChan)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-doneChan:
				return
			case <-ticker.C:
				quoteResp, err := getQuoteSell(tokenMint, lamports)
				if err != nil {
					continue
				}
				if outAmountStr, ok := quoteResp["outAmount"].(string); ok && outAmountStr != "0" {
					select {
					case quoteChan <- quoteResp:
					default:
					}
				}
			}
		}
	}()

	timeout := time.After(5 * time.Second)

	var finalQuoteResp map[string]interface{}
	var swapTx string
	var solReceived float64
	var err error

attemptLoop:
	for {
		select {
		case <-timeout:
			return tradeError("Sell failed: Timeout waiting for fresh quote", tokenMint, models.Sell, models.QUOTE_ERROR, currentSOLBalance)

		case quoteResp := <-quoteChan:
			finalQuoteResp = quoteResp

			swapTx, solReceived, err = getSwapTransaction(config.SOL_MINT, finalQuoteResp, 9)
			if err == nil {
				break attemptLoop
			}

			log.Printf("Swap preparation failed: %v. Retrying with fresh quote...\n", err)
		}
	}

	redisSuccess, originalSOL, originalToken, newSOL := operations.Sell(tokenMint, solReceived, finalSellAmount)
	if !redisSuccess {
		return tradeError("Failed to update Redis balances after sell", tokenMint, models.Sell, models.REDIS_ERROR, currentSOLBalance)
	}

	result := signAndSendTransaction(swapTx)
	success, ok := result["success"].(bool)
	if !ok || !success {
		rdb := database.GetRedisClient()
		ctx := database.GetRedisContext()

		_ = rdb.Set(ctx, "balance:sol", originalSOL, 0).Err()
		tokenKey := fmt.Sprintf("balance:token:%s", tokenMint)
		_ = rdb.Set(ctx, tokenKey, originalToken, 0).Err()

		return tradeError("Signing and sending transaction failed", tokenMint, models.Sell, models.SIGN_ERROR, currentSOLBalance)
	}

	return models.Trade{
		Success:      true,
		Timestamp:    time.Now(),
		TokenAddress: tokenMint,
		Side:         string(models.Sell),
		AmountSOL:    &solReceived,
		AmountToken:  &finalSellAmount,
		Balance:      &newSOL,
	}
}
