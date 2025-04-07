package helius

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/util"
)

func RawHelisusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Request")
	now := time.Now()
	fmt.Println(now.Format("2006-01-02 15:04:05.000"))

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST Requests", http.StatusMethodNotAllowed)
		return
	}

	// authHeader := r.Heamer.Get("Authorization")
	// expected := "Bearer " + config.BEARER_TOKEN

	// if authHeader != expected {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	var data []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(data) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	tx := data[0]

	meta, ok := tx["meta"].(map[string]interface{})
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	txTimeFloat, ok := tx["blockTime"].(float64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	txTime := int64(txTimeFloat)
	t := time.Unix(txTime, 0)
	fmt.Printf("Time for Processing Before: %f\n", time.Since(t).Seconds())

	preTokenBalances, _ := meta["preTokenBalances"].([]interface{})
	postTokenBalances, _ := meta["postTokenBalances"].([]interface{})
	preDict := make(map[string]float64)
	for _, preRaw := range preTokenBalances {
		pre, _ := preRaw.(map[string]interface{})
		if pre["owner"] == config.TRACKING_WALLET {
			key := fmt.Sprintf("%v-%v-%v", pre["accountIndex"], pre["mint"], pre["programId"])
			uiTokenAmount, _ := pre["uiTokenAmount"].(map[string]interface{})
			uiAmount, _ := uiTokenAmount["uiAmount"].(float64)
			preDict[key] = uiAmount
		}
	}
	postDict := make(map[string]float64)
	postInfo := make(map[string][2]interface{})
	for _, postRaw := range postTokenBalances {
		post, _ := postRaw.(map[string]interface{})
		if post["owner"] == config.TRACKING_WALLET {
			key := fmt.Sprintf("%v-%v-%v", post["accountIndex"], post["mint"], post["programId"])
			uiTokenAmount, _ := post["uiTokenAmount"].(map[string]interface{})
			uiAmount, _ := uiTokenAmount["uiAmount"].(float64)
			postDict[key] = uiAmount
			decimals := uiTokenAmount["decimals"]
			postInfo[key] = [2]interface{}{post["mint"], decimals}
		}
	}
	tokenChangeSum := 0.0
	tokenFound := false
	var tokenMint string
	var decimals interface{}
	var humanTokenAmount float64 = 0.0
	allKeys := make(map[string]struct{})
	for key := range preDict {
		allKeys[key] = struct{}{}
	}
	for key := range postDict {
		allKeys[key] = struct{}{}
	}
	for key := range allKeys {
		preAmount := preDict[key]
		postAmount := postDict[key]
		diff := postAmount - preAmount
		if diff != 0 {
			tokenFound = true
			tokenChangeSum += diff
			if info, ok := postInfo[key]; ok {
				tokenMint = info[0].(string)
				decimals = info[1]
			} else {
				parts := strings.Split(key, "-")
				if len(parts) >= 2 {
					tokenMint = parts[1]
				}
				decimals = 6
			}
			humanTokenAmount += math.Abs(diff)
		}
	}
	solDeltaSol := 0.0
	trans, _ := tx["transaction"].(map[string]interface{})
	message, _ := trans["message"].(map[string]interface{})
	accountKeysRaw, _ := message["accountKeys"].([]interface{})
	accountIndex := -1
	for idx, acc := range accountKeysRaw {
		if acc == config.TRACKING_WALLET {
			accountIndex = idx
			break
		}
	}

	preBalances, _ := meta["preBalances"].([]interface{})
	postBalances, _ := meta["postBalances"].([]interface{})
	if accountIndex >= 0 && accountIndex < len(preBalances) && accountIndex < len(postBalances) {
		preBal, _ := preBalances[accountIndex].(float64)
		postBal, _ := postBalances[accountIndex].(float64)
		if math.Abs(postBal-preBal) > 1 {
			solDeltaSol = (postBal - preBal) / 1e9
		}
	}

	feeLamports, _ := meta["fee"].(float64)
	feeSol := feeLamports / 1e9
	var txType string
	solAmount := 0.0
	if tokenFound {
		if tokenChangeSum < 0 {
			txType = "sell"
			solAmount = feeSol + solDeltaSol
		} else if tokenChangeSum > 0 {
			txType = "buy"
			solAmount = -solDeltaSol - feeSol
		} else {
			txType = "neutral"
			solAmount = math.Abs(solDeltaSol)
		}
	} else {
		if solDeltaSol > 0 {
			txType = "sell"
			solAmount = solDeltaSol
		} else if solDeltaSol < 0 {
			txType = "buy"
			solAmount = -solDeltaSol
		} else {
			txType = "neutral"
			solAmount = 0.0
		}
	}

	// var result models.Trade
	if txType == "buy" && len(tokenMint) > 0 && tokenMint != config.SOL_MINT {
		solAmount = util.CustomRound(solAmount)
		fmt.Printf("Time for Processing After: %f\n", time.Since(t).Seconds())
		fmt.Println("🟢 BUY Detected")
		fmt.Printf("  Mint: %v\n", tokenMint)
		fmt.Printf("  Amount: %.10f\n", humanTokenAmount)
		fmt.Printf("  Decimals: %v\n", decimals)
		fmt.Printf("  SOL Spent: %.10f\n", solAmount)

		// if time.Since(t) > 6*time.Second {
		// 	message := "Skipping stale BUY (>6s old)"
		// 	errStr := string(models.EXPIRED)
		// 	result = models.Trade{
		// 		Success:      false,
		// 		Timestamp:    time.Now(),
		// 		TokenAddress: tokenMint,
		// 		Side:         string(models.Buy),
		// 		AmountSOL:    nil,
		// 		AmountToken:  nil,
		// 		Message:      &message,
		// 		Error:        &errStr,
		// 	}
		// 	logging.LogTradeToCSV(result)
		// 	return
		// }

		// result = transaction.Buy(tokenMint, solAmount)
		// logging.LogTradeToCSV(result)
	} else if txType == "sell" && len(tokenMint) > 0 && tokenMint != config.SOL_MINT {
		fmt.Printf("Time for Processing After: %f\n", time.Since(t).Seconds())
		fmt.Println("🔴 SELL Detected")
		fmt.Printf("  Mint: %v\n", tokenMint)
		fmt.Printf("  Amount: %.10f\n", humanTokenAmount)
		fmt.Printf("  Decimals: %v\n", decimals)
		fmt.Printf("  SOL Received: %.10f\n", solAmount)

		// if time.Since(t) > 10*time.Second {
		// 	message := "Skipping stale SELL (>10s old)"
		// 	errStr := string(models.EXPIRED)
		// 	result = models.Trade{
		// 		Success:      false,
		// 		Timestamp:    time.Now(),
		// 		TokenAddress: tokenMint,
		// 		Side:         string(models.Sell),
		// 		AmountSOL:    nil,
		// 		AmountToken:  nil,
		// 		Message:      &message,
		// 		Error:        &errStr,
		// 	}
		// 	logging.LogTradeToCSV(result)

		// }

		// result = transaction.Sell(tokenMint, solAmount)
		// logging.LogTradeToCSV(result)

	} else {
		fmt.Println("Neutral or unclear transaction type.")
	}
	now = time.Now()
	fmt.Println(now.Format("2006-01-02 15:04:05.000"))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}
