package main

import (
	"log"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/database"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/logging"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/transaction"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helloWorld"
)

func main() {

	var success bool

	success = database.InitializeRedis()
	if !success {
		log.Fatal("Redis is not connected")
	}

	log.Println("Redis is connected!")

	success = database.DeleteAllRedis()
	if !success {
		log.Fatal("Not able to delete all in Redis")
	}

	logging.StartCSVLogger(100)
	defer logging.StopCSVLogger()

	transaction.InitKeypair()
	http.HandleFunc("/", helloWorld.HelloWorldHandler)
	http.HandleFunc("/helius", helius.RawHelisusHandler)

	log.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
