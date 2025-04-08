package helloWorld

import (
	"fmt"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/transaction"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	mint := r.URL.Query().Get("mint")
	if mint == "" {
		http.Error(w, "mint parameter is required", http.StatusBadRequest)
		return
	}

	//Use the provided mint for the Buy transaction
	result := transaction.Buy(mint, 0.5, 6)
	if result.Success {
		fmt.Fprintf(w, "Buy transaction went through\n")
		fmt.Fprintf(w, "Amount of token bought: %f\n", *result.AmountToken)
		result1 := transaction.Sell(mint, *result.AmountToken*config.POSITION_SIZE, 6)
		if result1.Success {
			fmt.Fprintf(w, "Buy and Sell transactions went through")
		} else {
			fmt.Fprintf(w, "Buy transaction worked but Sell didn't")
		}
	} else {
		fmt.Fprintf(w, "Neither Buy nor Sell transactions worked")
	}

	// result := transaction.Sell(mint, 847.4, 6)
	// if result.Success {
	// 	fmt.Fprintf(w, "Selling Works")
	// } else {
	// 	fmt.Fprintf(w, "Selling did not work")
	// }
}
