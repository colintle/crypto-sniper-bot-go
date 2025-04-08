package util

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
)

func GetSolBalance(ctx context.Context) float64 {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params":  []interface{}{config.PUBLIC_KEY_WALLET},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return 0
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.HELIUS_URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return 0
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to perform request: %v", err)
		return 0
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return 0
	}

	result, ok := res["result"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid response format: missing 'result'")
		return 0
	}

	value, ok := result["value"].(float64)
	if !ok {
		log.Printf("Invalid 'value' type in result")
		return 0
	}

	return value / config.LAMPORTS_PER_SOL
}

func GetTokenBalance(ctx context.Context, tokenAddress string) float64 {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenAccountsByOwner",
		"params": []interface{}{
			config.PUBLIC_KEY_WALLET,
			map[string]interface{}{"mint": tokenAddress},
			map[string]interface{}{"encoding": "jsonParsed"},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return 0
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.HELIUS_URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return 0
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return 0
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Printf("Error decoding response: %v", err)
		return 0
	}

	result, ok := res["result"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid response format (missing result)")
		return 0
	}

	values, ok := result["value"].([]interface{})
	if !ok {
		log.Printf("Invalid response format (missing value list)")
		return 0
	}

	var totalBalance float64

	for _, v := range values {
		account, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		accountData, ok := account["account"].(map[string]interface{})
		if !ok {
			continue
		}
		data, ok := accountData["data"].(map[string]interface{})
		if !ok {
			continue
		}
		parsed, ok := data["parsed"].(map[string]interface{})
		if !ok {
			continue
		}
		info, ok := parsed["info"].(map[string]interface{})
		if !ok {
			continue
		}
		tokenAmount, ok := info["tokenAmount"].(map[string]interface{})
		if !ok {
			continue
		}
		uiAmount, ok := tokenAmount["uiAmount"].(float64)
		if !ok {
			continue
		}
		totalBalance += uiAmount
	}

	return totalBalance
}
