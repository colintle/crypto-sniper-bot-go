package models

type TokenAmount struct {
	UiAmount float64 `json:"uiAmount"`
	Decimals int     `json:"decimals"`
}

type TokenBalance struct {
	AccountIndex   int         `json:"accountIndex"`
	Mint           string      `json:"mint"`
	Owner          string      `json:"owner"`
	ProgramId      string      `json:"programId"`
	UiTokenAmount  TokenAmount `json:"uiTokenAmount"`
}

type Meta struct {
	PreTokenBalances  []TokenBalance `json:"preTokenBalances"`
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`
	PreBalances       []float64      `json:"preBalances"`
	PostBalances      []float64      `json:"postBalances"`
	Fee               float64        `json:"fee"`
}

type Transaction struct {
	Message struct {
		AccountKeys []string `json:"accountKeys"`
	} `json:"message"`
}

type TxData struct {
	Meta        Meta        `json:"meta"`
	BlockTime   int64       `json:"blockTime"`
	Transaction Transaction `json:"transaction"`
}

type QuoteResponse struct {
	OutAmount string `json:"outAmount"`
	Route     Route  `json:"route"`
}

type Route struct {
	InAmount          string           `json:"inAmount"`
	OutAmount         string           `json:"outAmount"`
}
