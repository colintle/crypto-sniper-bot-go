package helius

import (
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/logging"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/transaction"
	"github.com/colintle/crypto-sniper-bot-go/models"
)

func EnhancedHeliusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST Requests", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	expected := "Bearer " + config.BEARER_TOKEN

	if authHeader != expected {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// currentTime := time.Now()
	// fmt.Printf("Received Timestamp from Webhook: %s\n", currentTime.Format("2006-01-02 15:04:05"))

	for _, tx := range data {
		txTimestampRaw, _ := tx["timestamp"].(float64)
		txTime := time.Unix(int64(txTimestampRaw), 0)

		// fmt.Printf("Transaction Timestamp: %s\n", txTime.Format("2006-01-02 15:04:05"))

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

		var result models.Trade
		for _, t := range tokens {
			token := t.(map[string]interface{})

			amount, _ := token["tokenAmount"].(float64)
			mint, _ := token["mint"].(string)

			toUser := token["toUserAccount"]
			fromUser := token["fromUserAccount"]

			if toUser == config.TRACKING_WALLET {
				if time.Since(txTime) > 6*time.Second {
					message := "Skipping stale BUY (>6s old)"
					errStr := string(models.EXPIRED)
					result = models.Trade{
						Success:      false,
						Timestamp:    time.Now(),
						TokenAddress: mint,
						Side:         string(models.Buy),
						AmountSOL:    nil,
						AmountToken:  nil,
						Message:      &message,
						Error:        &errStr,
					}
					logging.LogTradeToCSV(result)
					continue
				}

				if mint == config.SOL_MINT {
					message := "Case in trader is buying SOL"
					errStr := string(models.BUY_SOL)
					result = models.Trade{
						Success:      false,
						Timestamp:    time.Now(),
						TokenAddress: mint,
						Side:         string(models.Buy),
						AmountSOL:    nil,
						AmountToken:  nil,
						Message:      &message,
						Error:        &errStr,
					}
					logging.LogTradeToCSV(result)
					continue
				}

				solSpent := 0.0
				if solChange < 0 {
					solSpent = -solChange
				}
				// price := 0.0
				// if amount != 0 {
				// 	price = solSpent / amount
				// }

				// fmt.Println("🟢 BUY Detected")
				// fmt.Printf("  Mint: %s\n", mint)
				// fmt.Printf("  Amount: %f\n", amount)
				// fmt.Printf("  SOL Spent: %.10f\n", solSpent)
				// fmt.Printf("  Price per token: %.10f SOL\n\n", price)

				result = transaction.Buy(mint, solSpent, 6)
				logging.LogTradeToCSV(result)

			} else if fromUser == config.TRACKING_WALLET {
				if time.Since(txTime) > 10*time.Second {
					message := "Skipping stale SELL (>10s old)"
					errStr := string(models.EXPIRED)
					result = models.Trade{
						Success:      false,
						Timestamp:    time.Now(),
						TokenAddress: mint,
						Side:         string(models.Sell),
						AmountSOL:    nil,
						AmountToken:  nil,
						Message:      &message,
						Error:        &errStr,
					}
					logging.LogTradeToCSV(result)
					continue
				}
				// solReceived := 0.0
				// if solChange > 0 {
				// 	solReceived = solChange
				// }
				// price := 0.0
				// if amount != 0 {
				// 	price = solReceived / amount
				// }

				// fmt.Println("🔴 SELL Detected")
				// fmt.Printf("  Mint: %s\n", mint)
				// fmt.Printf("  Amount: %f\n", amount)
				// fmt.Printf("  SOL Received: %.10f\n", solReceived)
				// fmt.Printf("  Price per token: %.10f SOL\n\n", price)

				result = transaction.Sell(mint, amount, 6)
				logging.LogTradeToCSV(result)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}
