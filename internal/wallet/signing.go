package wallet

import (
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func (w Wallet) SignTransaction(transaction *tx.Transaction) error {
	return tx.SignTransaction(transaction, w.PrivateKey)
}
