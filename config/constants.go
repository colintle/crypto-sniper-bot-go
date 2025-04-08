package config

import "os"

const (
	SELL_PRIORITY_FEE = 100000000 // Higher for sells

	POSITION_SIZE     = 10
	MAX_POSITION      = 0.5
	MIN_BALANCE       = 2
	JUPITER_QUOTE_URL = "https://api.jup.ag/swap/v1/quote"
	JUPITER_SWAP_URL  = "https://api.jup.ag/swap/v1/swap"

	SOLANA_RPC_URL   = "https://api.mainnet-beta.solana.com"
	SOL_MINT         = "So11111111111111111111111111111111111111112"
	LAMPORTS_PER_SOL = 1_000_000_000
)

var (
	PUBLIC_KEY_WALLET  = os.Getenv("PUBLIC_KEY_WALLET")
	BASE58_PRIVATE_KEY = os.Getenv("PRIVATE_KEY_WALLET")
	TRACKING_WALLET    = os.Getenv("TRACKING_WALLET")
	API_KEY            = os.Getenv("HELIUS_API_KEY")
	HELIUS_URL         = os.Getenv("HELIUS_URL")
	BEARER_TOKEN       = os.Getenv("BEARER_TOKEN")
	JUPITER_API_KEY    = os.Getenv("JUPITER_API_KEY")
)
