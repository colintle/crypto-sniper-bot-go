package operations

import (
	"fmt"
	"log"

	"github.com/colintle/crypto-sniper-bot-go/database"
)

func Buy(tokenAddress string, solSpent, tokenReceived float64) (bool, float64, float64, float64) {
	rdb := database.GetRedisClient()
	ctx := database.GetRedisContext()

	originalSOL := GetSOLBalanceCached()
	originalToken := GetTokenBalanceCached(tokenAddress)

	newSOL := originalSOL - solSpent
	newToken := originalToken + tokenReceived

	if err := rdb.Set(ctx, "balance:sol", newSOL, 0).Err(); err != nil {
		log.Printf("Failed to update SOL balance after buy: %v", err)
		return false, originalSOL, originalToken, originalSOL
	}

	tokenKey := fmt.Sprintf("balance:token:%s", tokenAddress)
	if err := rdb.Set(ctx, tokenKey, newToken, 0).Err(); err != nil {
		log.Printf("Failed to update token balance after buy: %v", err)
		_ = rdb.Set(ctx, "balance:sol", originalSOL, 0).Err()
		return false, originalSOL, originalToken, originalSOL
	}

	return true, originalSOL, originalToken, newSOL
}

func Sell(tokenAddress string, solReceived, tokenSpent float64) (bool, float64, float64, float64) {
	rdb := database.GetRedisClient()
	ctx := database.GetRedisContext()

	originalSOL := GetSOLBalanceCached()
	originalToken := GetTokenBalanceCached(tokenAddress)

	newSOL := originalSOL + solReceived
	if err := rdb.Set(ctx, "balance:sol", newSOL, 0).Err(); err != nil {
		log.Printf("Failed to update SOL balance after sell: %v", err)
		return false, originalSOL, originalToken, originalSOL
	}

	newToken := originalToken - tokenSpent
	if newToken < 0 {
		newToken = 0
	}

	tokenKey := fmt.Sprintf("balance:token:%s", tokenAddress)

	if newToken < 0.01 {
		if err := rdb.Del(ctx, tokenKey).Err(); err != nil {
			log.Printf("Failed to delete token key after sell: %v", err)
			_ = rdb.Set(ctx, "balance:sol", originalSOL, 0).Err()
			return false, originalSOL, originalToken, originalSOL
		}
		log.Printf("Deleted token key: %s (balance = 0)", tokenKey)
	} else {
		if err := rdb.Set(ctx, tokenKey, newToken, 0).Err(); err != nil {
			log.Printf("Failed to update token balance after sell: %v", err)
			_ = rdb.Set(ctx, "balance:sol", originalSOL, 0).Err()
			return false, originalSOL, originalToken, originalSOL
		}
	}

	return true, originalSOL, originalToken, newSOL
}
