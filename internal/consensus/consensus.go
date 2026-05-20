package consensus

import (
	"crypto/ed25519"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type Engine interface {
	BuildAndSign(parent chain.Block, transactions []tx.Transaction) (chain.Block, error)
	Validate(block chain.Block) error
}

type Config struct {
	ValidatorAddress string
	ValidatorKey     ed25519.PrivateKey
	Validators       map[string]bool
	ValidatorList    []string
}
