package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/models"
	"github.com/colintle/crypto-sniper-bot-go/util"
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

func getSwapTransaction(tokenMint string, quoteResp map[string]interface{}) (string, float64, error) {
	swapReq := map[string]interface{}{
		"userPublicKey": WalletPublicKey.String(),
		"quoteResponse": quoteResp,
	}
	payload, _ := json.Marshal(swapReq)

	resp, err := http.Post(config.JUPITER_SWAP_URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var swapData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&swapData); err != nil {
		return "", 0, err
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

	decimals := util.GetTokenDecimals(tokenMint)
	if decimals == nil {
		return "", 0, fmt.Errorf("missing decimals")
	}

	trueToken := outAmountParsed / float64(intPow(10, *decimals))
	return tx, trueToken, nil
}

func getQuote(tokenMint string, lamports int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d",
		config.JUPITER_QUOTE_URL, config.SOL_MINT, tokenMint, lamports, config.SLIPPAGE)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var quoteResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, err
	}
	if _, ok := quoteResp["routePlan"]; !ok {
		return nil, fmt.Errorf("no route to purchase token")
	}
	return quoteResp, nil
}

func getQuoteSell(tokenMint string, lamports int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d",
		config.JUPITER_QUOTE_URL, tokenMint, config.SOL_MINT, lamports, config.SLIPPAGE)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var quoteResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, err
	}
	if _, ok := quoteResp["routePlan"]; !ok {
		return nil, fmt.Errorf("no route to sell token")
	}
	return quoteResp, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
