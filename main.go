package main

import (
	"log"
	"net/http"

	// "github.com/colintle/crypto-sniper-bot-go/handlers/helius"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helloWorld"
	"github.com/colintle/crypto-sniper-bot-go/database"
)

func main(){

	success := database.InitializeRedis()

	if !success{
		log.Fatal("Redis is not connected")
	}

	log.Println("Redis is connected!")

	http.HandleFunc("/", helloWorld.HelloWorldHandler)
	//http.HandleFunc("/helius", helius.HeliusHandler)

	log.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
