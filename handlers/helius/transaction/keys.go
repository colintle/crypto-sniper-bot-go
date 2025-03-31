package transaction

import (
	"log"

	"github.com/colintle/crypto-sniper-bot-go/config"
	"github.com/gagliardetto/solana-go"
	"github.com/jbenet/go-base58"
)

var WalletPrivateKey solana.PrivateKey
var WalletPublicKey solana.PublicKey

func InitKeypair() {
	secretKeyBytes := base58.Decode(config.BASE58_PRIVATE_KEY)
	if len(secretKeyBytes) != 64 {
		log.Fatalf("Invalid secret key length: expected 64 bytes, got %d", len(secretKeyBytes))
	}

	WalletPrivateKey = solana.PrivateKey(secretKeyBytes)
	WalletPublicKey = WalletPrivateKey.PublicKey()
}
