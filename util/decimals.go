package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/config"
)

func GetTokenDecimals(tokenMint string) *int {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenSupply",
		"params":  []interface{}{tokenMint},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return nil
	}

	resp, err := http.Post(config.SOLANA_RPC_URL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("RPC error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Decode error: %v", err)
		return nil
	}

	if val, ok := result["result"].(map[string]interface{}); ok {
		if supply, ok := val["value"].(map[string]interface{}); ok {
			if dec, ok := supply["decimals"].(float64); ok {
				d := int(dec)
				return &d
			}
		}
	}
	log.Printf("Unexpected structure: %+v", result)
	return nil
}
