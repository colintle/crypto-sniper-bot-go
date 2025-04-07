package helloWorld

import (
	"fmt"
	"net/http"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/colintle/crypto-sniper-bot-go/handlers/helius/transaction"
	"github.com/colintle/crypto-sniper-bot-go/models"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {

	// fmt.Println("Testing")
	// fmt.Fprintf(w, "Hello World!")

	var result models.Trade = transaction.Buy("7HoCDyzPwSfgrPxYCA5w1XetujmBtDhUPUnTrEF1pump", 0.05)
	if result.Success {
		fmt.Fprintf(w, "Transaction went through")
		result1 := transaction.Sell("7HoCDyzPwSfgrPxYCA5w1XetujmBtDhUPUnTrEF1pump", *result.AmountToken*config.POSITION_SIZE, 6)
		if result1.Success {
			fmt.Fprintf(w, "Buy and Sell went through")
		} else {
			fmt.Fprintf(w, "Buy works but sell didn't")
		}
	} else {
		fmt.Fprintf(w, "Neither buy or sell works")
	}
}
