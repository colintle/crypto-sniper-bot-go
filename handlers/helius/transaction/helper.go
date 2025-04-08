package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/models"
)

func tradeError(msg string, tokenMint string, side models.Side, error models.Error, balance float64) models.Trade {
	log.Println(msg)
	errStr := string(error)
	return models.Trade{
		Success:      false,
		Timestamp:    time.Now(),
		TokenAddress: tokenMint,
		Side:         string(side),
		Message:      &msg,
		Balance:      &balance,
		Error:        &errStr,
	}
}

func getSwapTransaction(tokenMint string, quoteResp map[string]interface{}, decimals int) (string, float64, error) {
	swapReq := map[string]interface{}{
		"userPublicKey": WalletPublicKey.String(),
		"quoteResponse": quoteResp,
		"prioritizationFeeLamports": map[string]interface{}{
			"priorityLevelWithMaxLamports": map[string]interface{}{
				"maxLamports":   config.SELL_PRIORITY_FEE,
				"priorityLevel": "veryHigh",
			},
		},
		"dynamicComputeUnitLimit": true,
		"dynamicSlippage": true,
	}

	payload, _ := json.Marshal(swapReq)

	req, err := http.NewRequest("POST", config.JUPITER_SWAP_URL, bytes.NewBuffer(payload))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.JUPITER_API_KEY)

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var swapData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&swapData); err != nil {
		return "", 0, err
	}

	// swapDataBytes, err := json.MarshalIndent(swapData, "", "  ")
	// if err != nil {
	// 	fmt.Printf("Error marshaling swapData for printing: %v\n", err)
	// } else {
	// 	fmt.Printf("🔍 Swap response:\n%s\n", string(swapDataBytes))
	// }

	if simErr, ok := swapData["simulationError"]; ok && simErr != nil {
		// simErrMap, isMap := simErr.(map[string]interface{})
		// if isMap {
		// 	fmt.Printf("❌ Simulation error detected: %v\n", simErrMap["error"])
		// } else {
		// 	fmt.Printf("❌ Simulation error detected: %v\n", simErr)
		// }
		return "", 0, fmt.Errorf("swap simulation error: %v", simErr)
	}

	tx, ok := swapData["swapTransaction"].(string)
	if !ok {
		return "", 0, fmt.Errorf("missing swapTransaction field")
	}

	outAmountStr, ok := quoteResp["outAmount"].(string)
	if !ok {
		return "", 0, fmt.Errorf("outAmount not a string")
	}
	outAmountParsed, err := strconv.ParseFloat(outAmountStr, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid outAmount: %v", err)
	}

	trueToken := outAmountParsed / float64(intPow(10, decimals))
	return tx, trueToken, nil
}

func getQuote(tokenMint string, lamports int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&dynamicSlippage=true",
		config.JUPITER_QUOTE_URL,
		config.SOL_MINT,
		tokenMint,
		lamports,
	)

	// fmt.Printf("📤 Requesting Quote URL: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", config.JUPITER_API_KEY)

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// var prettyJSON bytes.Buffer
	// if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
	// 	fmt.Printf("⚠️ Failed to format JSON: %v\n", err)
	// 	fmt.Printf("Raw body: %s\n", string(bodyBytes))
	// } else {
	// 	fmt.Printf("🔍 Quote Response:\n%s\n", prettyJSON.String())
	// }

	var quoteResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &quoteResp); err != nil {
		return nil, err
	}

	if _, ok := quoteResp["routePlan"]; !ok {
		return nil, fmt.Errorf("no route to purchase token")
	}

	outAmountStr, ok := quoteResp["outAmount"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid outAmount in quote")
	}
	outAmount, err := strconv.ParseFloat(outAmountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid outAmount value")
	}
	if outAmount == 0 {
		return nil, fmt.Errorf("outAmount is zero, no liquidity")
	}

	return quoteResp, nil
}

func getQuoteSell(tokenMint string, lamports int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&dynamicSlippage=true",
		config.JUPITER_QUOTE_URL,
		tokenMint,
		config.SOL_MINT,
		lamports,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", config.JUPITER_API_KEY)

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// var prettyJSON bytes.Buffer
	// if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
	// 	fmt.Printf("⚠️ Failed to format JSON: %v\n", err)
	// 	fmt.Printf("Raw body: %s\n", string(bodyBytes))
	// } else {
	// 	fmt.Printf("🔍 Sell Quote Response:\n%s\n", prettyJSON.String())
	// }

	var quoteResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &quoteResp); err != nil {
		return nil, err
	}

	if _, ok := quoteResp["routePlan"]; !ok {
		return nil, fmt.Errorf("no route to sell token")
	}

	outAmountStr, ok := quoteResp["outAmount"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid outAmount in quote")
	}
	outAmount, err := strconv.ParseFloat(outAmountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid outAmount value")
	}
	if outAmount == 0 {
		return nil, fmt.Errorf("outAmount is zero, no liquidity")
	}

	return quoteResp, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
