package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/config"
)

func GetSolBalance() float64 {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params":  []interface{}{config.PUBLIC_KEY_WALLET},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(config.SOLANA_RPC_URL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatal(err)
	}

	result := res["result"].(map[string]interface{})
	lamports := result["value"].(float64)
	sol := lamports / config.LAMPORTS_PER_SOL

	return sol
}

func GetTokenBalance (tokenAddress string) float64{
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
	if err != nil{
		log.Fatal(err)
	}

	resp, err := http.Post(config.SOLANA_RPC_URL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil{
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil{
		log.Fatal(err)
	}

	result := res["result"].(map[string]interface{})

	value := result["value"].([]interface{})
	var totalBalance float64

	for _, account := range value {
		accountMap := account.(map[string]interface{})
		accountData := accountMap["account"].(map[string]interface{})
		data := accountData["data"].(map[string]interface{})
		parsed := data["parsed"].(map[string]interface{})
		info := parsed["info"].(map[string]interface{})
		tokenAmount := info["tokenAmount"].(map[string]interface{})
	
		uiAmountRaw, ok := tokenAmount["uiAmount"]
		if !ok {
			continue
		}
	
		uiAmount, ok := uiAmountRaw.(float64)
		if !ok {
			continue
		}
	
		totalBalance += uiAmount
	}

	return totalBalance
}
