package transaction

import (
	"fmt"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/database"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/operations"
	"github.com/colintle/crypto-sniper-bot-go/models"
	"github.com/colintle/crypto-sniper-bot-go/util"
)

func Sell(tokenMint string, tokenAmount float64) models.Trade {
	currentBalance := operations.GetTokenBalanceCached(tokenMint)
	currentSOLBalance := operations.GetSOLBalanceCached()
	if currentBalance == 0 {
		return tradeError("No tokens to sell; balance is 0.", tokenMint, models.Sell, models.LOW_TOKEN, currentSOLBalance)
	}

	portionToSell := tokenAmount / config.POSITION_SIZE
	finalSellAmount := min(portionToSell, currentBalance)
	if finalSellAmount <= 0 {
		return tradeError("No tokens to sell after final calculation.", tokenMint, models.Sell, models.MIN_TOKEN, currentSOLBalance)
	}

	decimals := util.GetTokenDecimals(tokenMint)
	if decimals == nil {
		return tradeError("Missing decimals", tokenMint, models.Sell, models.MISSING_DEC, currentSOLBalance)
	}

	lamports := int(finalSellAmount * float64(intPow(10, *decimals)))
	quoteResp, err := getQuoteSell(tokenMint, lamports)
	if err != nil {
		return tradeError(fmt.Sprintf("Quote error: %v", err), tokenMint, models.Sell, models.QUOTE_ERROR, currentSOLBalance)
	}

	// converts the outAmount from lamports to not lamports
	swapTx, solReceived, err := getSwapTransaction(config.SOL_MINT, quoteResp)
	if err != nil {
		return tradeError(fmt.Sprintf("Swap preparation failed: %v", err), tokenMint, models.Sell, models.SWAP_ERROR, currentSOLBalance)
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
