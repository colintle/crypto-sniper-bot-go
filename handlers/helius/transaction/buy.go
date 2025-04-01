package transaction

import (
	"fmt"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/database"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/operations"
	"github.com/colintle/crypto-sniper-bot-go/models"
)

func Buy(tokenMint string, solSpent float64) models.Trade {
	buySolAmount := solSpent / config.POSITION_SIZE
	balance := operations.GetSOLBalanceCached()

	if balance < config.MIN_BALANCE {
		return tradeError("Not enough SOL", tokenMint, models.Buy, models.MINIMUM_SOL, balance)
	}

	if config.MAX_POSITION < buySolAmount {
		return tradeError("Buy SOL amount is bigger than max position", tokenMint, models.Buy, models.LOW_SOL, balance)
	}

	lamports := int(buySolAmount * config.LAMPORTS_PER_SOL)
	quoteResp, err := getQuote(tokenMint, lamports)
	if err != nil {
		return tradeError(fmt.Sprintf("Quote error: %v", err), tokenMint, models.Buy, models.QUOTE_ERROR, balance)
	}

	// converts the outAmount from lamports to not lamports
	swapTx, trueToken, err := getSwapTransaction(tokenMint, quoteResp)
	if err != nil {
		return tradeError(fmt.Sprintf("Swap preparation failed: %v", err), tokenMint, models.Buy, models.SWAP_ERROR, balance)
	}

	redisSuccess, originalSOL, originalToken, newSOL := operations.Buy(tokenMint, buySolAmount, trueToken)
	if !redisSuccess {
		return tradeError("Failed to insert to Redis", tokenMint, models.Buy, models.REDIS_ERROR, balance)
	}

	result := signAndSendTransaction(swapTx)
	success, ok := result["success"].(bool)
	if !ok || !success {
		rdb := database.GetRedisClient()
		ctx := database.GetRedisContext()

		rdb.Set(ctx, "balance:sol", originalSOL, 0)
		tokenKey := fmt.Sprintf("balance:token:%s", tokenMint)
		rdb.Set(ctx, tokenKey, originalToken, 0)

		return tradeError("Signing and sending transaction failed", tokenMint, models.Buy, models.SIGN_ERROR, balance)
	}

	return models.Trade{
		Success:      true,
		Timestamp:    time.Now(),
		TokenAddress: tokenMint,
		Side:         string(models.Buy),
		AmountSOL:    &buySolAmount,
		AmountToken:  &trueToken,
		Balance:      &newSOL,
	}
}
