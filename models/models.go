package models

import (
	"time"
)

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type Error string

const (
	EXPIRED     Error = "expired"
	BUY_SOL     Error = "buy_sol"
	MINIMUM_SOL Error = "minimum_sol"
	LOW_SOL     Error = "low_sol"
	QUOTE_ERROR Error = "quote_error"
	SWAP_ERROR  Error = "swap_error"
	REDIS_ERROR Error = "redis_error"
	SIGN_ERROR  Error = "sign_error"
	LOW_TOKEN   Error = "low_token"
	MIN_TOKEN   Error = "min_token"
	MISSING_DEC Error = "missing_dec"
)

type Trade struct {
	Success      bool
	Timestamp    time.Time
	TokenAddress string
	Side         string
	AmountSOL    *float64
	AmountToken  *float64
	Balance      *float64
	Message      *string
	Error        *string
}
