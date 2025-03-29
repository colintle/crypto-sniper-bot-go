package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/go-redis/redis/v8"
	"context"
)

var ctx = context.Background()

func main(){
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		err := rdb.Set(ctx, "key", "Hello Redis!", 0).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := rdb.Get(ctx, "key").Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Redis says: %s\n", val)

	})

	log.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

