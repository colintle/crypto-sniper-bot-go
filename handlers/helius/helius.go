package helius

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/config"
)

func HeliusHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Only POST Requests", http.StatusMethodNotAllowed)
		return
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Fatal(err)
		return
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Received Timestamp from Webhook: %s\n", currentTime)
	
	for _, tx := range data{
		txTimestampRaw, _ := tx["timestamp"].(float64)
		txTime := time.Unix(int64(txTimestampRaw), 0)

		fmt.Printf("🧾 Transaction Timestamp: %s\n", txTime.Format("2006-01-02 15:04:05"))

		tokens, _ := tx["tokenTransfers"].([]interface{})
		accounts, _ := tx["accountData"].([]interface{})

		solChange := 0.0
		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			if account["account"] == config.TRACKING_WALLET {
				if val, ok := account["nativeBalanceChange"].(float64); ok {
					solChange = val / float64(config.LAMPORTS_PER_SOL)
					break
				}
			}
		}

		for _, t := range tokens {
			token := t.(map[string]interface{})

			amount, _ := token["tokenAmount"].(float64)
			mint, _ := token["mint"].(string)

			toUser := token["toUserAccount"]
			fromUser := token["fromUserAccount"]

			if toUser == config.TRACKING_WALLET {
				solSpent := 0.0
				if solChange < 0 {
					solSpent = -solChange
				}
				price := 0.0
				if amount != 0 {
					price = solSpent / amount
				}

				fmt.Println("🟢 BUY Detected")
				fmt.Printf("  Mint: %s\n", mint)
				fmt.Printf("  Amount: %f\n", amount)
				fmt.Printf("  SOL Spent: %.10f\n", solSpent)
				fmt.Printf("  Price per token: %.10f SOL\n\n", price)

			} else if fromUser == config.TRACKING_WALLET {
				solReceived := 0.0
				if solChange > 0 {
					solReceived = solChange
				}
				price := 0.0
				if amount != 0 {
					price = solReceived / amount
				}

				fmt.Println("🔴 SELL Detected")
				fmt.Printf("  Mint: %s\n", mint)
				fmt.Printf("  Amount: %f\n", amount)
				fmt.Printf("  SOL Received: %.10f\n", solReceived)
				fmt.Printf("  Price per token: %.10f SOL\n\n", price)

			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}
