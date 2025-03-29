package database

import (
	"context"
	"log"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func GetRedisClient() *redis.Client{
	return rdb
}

func GetRedisContext() context.Context{
	return ctx
}

func InitializeRedis() bool{
	rdb = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	_, err := rdb.Ping(ctx).Result()

	if err != nil{
		log.Fatal("Couldn't connect to Redis: %v", err)
		return false
	}

	fmt.Println("Connected to Redis")
	return true
}
