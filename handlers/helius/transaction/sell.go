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

	redisSuccess, originalSOL, originalToken, newSOL := operations.Sell(tokenMint, currentSOLBalance, finalSellAmount)
	if !redisSuccess {
		return tradeError("Failed to update Redis balances before sell", tokenMint, models.Sell, models.REDIS_ERROR, currentSOLBalance)
	}

	rdb := database.GetRedisClient()
	ctx := database.GetRedisContext()

	doneChan := make(chan models.Trade)

	go func() {
		defer close(doneChan)

		for {
			quoteResp, err := getQuoteSell(tokenMint, lamports)
			if err != nil {
				// log.Println("Sell: quote fetch error:", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			swapTx, solReceived, err := getSwapTransaction(config.SOL_MINT, quoteResp, 9)
			if err != nil {
				// log.Println("Sell: swap transaction prep error:", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// log.Println("Sell: got valid swap transaction")

			result := signAndSendTransaction(swapTx)
			success, ok := result["success"].(bool)
			if ok && success {
				log.Println("Sell: signed and sent transaction successfully")
				doneChan <- models.Trade{
					Success:      true,
					Timestamp:    time.Now(),
					TokenAddress: tokenMint,
					Side:         string(models.Sell),
					AmountSOL:    &solReceived,
					AmountToken:  &finalSellAmount,
					Balance:      &newSOL,
				}
				return
			}

			// log.Printf("Sell: sign and send failed: %v. Retrying...\n", result["error"])
			time.Sleep(500 * time.Millisecond)
		}
	}()

	result, ok := <-doneChan
	if !ok || !result.Success {
		log.Println("Sell: transaction failed, rolling back Redis balances")
		_ = rdb.Set(ctx, "balance:sol", originalSOL, 0).Err()
		tokenKey := fmt.Sprintf("balance:token:%s", tokenMint)
		_ = rdb.Set(ctx, tokenKey, originalToken, 0).Err()

		return tradeError("Sell: transaction failed after retries", tokenMint, models.Sell, models.SIGN_ERROR, currentSOLBalance)
	}
	return result
}
