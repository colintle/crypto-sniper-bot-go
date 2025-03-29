package main

import (
	"log"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/handlers"
	"github.com/colintle/crypto-sniper-bot-go/database"
)

func main(){

	success := database.InitializeRedis()

	if !success{
		log.Fatal("Redis is not connected")
	}

	log.Println("Redis is connected!")

	http.HandleFunc("/", handlers.HelloWorldHandler)
	http.HandleFunc("/helius", handlers.HeliusHandler)

	log.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

