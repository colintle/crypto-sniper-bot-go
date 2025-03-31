package operations

import (
	"fmt"
	"log"
	"strconv"

	"github.com/colintle/crypto-sniper-bot-go/database"
	"github.com/colintle/crypto-sniper-bot-go/util"
	"github.com/go-redis/redis/v8"
)

func GetSOLBalanceCached() float64 {
	rdb := database.GetRedisClient()
	ctx := database.GetRedisContext()

	key := "balance:sol"
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		balance := util.GetSolBalance(ctx)
		err := rdb.Set(ctx, key, balance, 0).Err()
		if err != nil {
			log.Printf("Failed to cache SOL balance: %v", err)
		}
		return balance
	} else if err != nil {
		log.Printf("Redis error: %v", err)
		return 0
	}

	balance, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Error parsing SOL balance: %v", err)
		return 0
	}

	return balance
}

func GetTokenBalanceCached(tokenAddress string) float64 {
	rdb := database.GetRedisClient()
	ctx := database.GetRedisContext()

	key := fmt.Sprintf("balance:token:%s", tokenAddress)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		balance := util.GetTokenBalance(ctx, tokenAddress)
		err := rdb.Set(ctx, key, balance, 0).Err()
		if err != nil {
			log.Printf("Failed to cache token balance: %v", err)
		}
		return balance
	} else if err != nil {
		log.Printf("Redis error: %v", err)
		return 0
	}

	balance, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Error parsing token balance: %v", err)
		return 0
	}

	return balance
}
