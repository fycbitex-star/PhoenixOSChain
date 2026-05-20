package tx

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
)

type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	Nonce     uint64 `json:"nonce"`
	GasLimit  uint64 `json:"gas_limit"`
	GasPrice  uint64 `json:"gas_price"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
	Hash      string `json:"hash"`
}

type unsignedTransaction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   uint64 `json:"amount"`
	Nonce    uint64 `json:"nonce"`
	GasLimit uint64 `json:"gas_limit"`
	GasPrice uint64 `json:"gas_price"`
}

func New(from, to string, amount, nonce, gasLimit, gasPrice uint64) Transaction {
	return Transaction{
		From:     from,
		To:       to,
		Amount:   amount,
		Nonce:    nonce,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}
}

func (t Transaction) UnsignedPayload() []byte {
	payload, _ := json.Marshal(unsignedTransaction{
		From:     t.From,
		To:       t.To,
		Amount:   t.Amount,
		Nonce:    t.Nonce,
		GasLimit: t.GasLimit,
		GasPrice: t.GasPrice,
	})
	return payload
}

func (t Transaction) ComputeHash() string {
	return phxcrypto.HashBytes(t.UnsignedPayload())
}

func (t *Transaction) RefreshHash() {
	t.Hash = t.ComputeHash()
}

func (t Transaction) VerifySignature() error {
	if t.From == "" || t.To == "" {
		return fmt.Errorf("from and to are required")
	}
	if t.Amount == 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if t.Hash != t.ComputeHash() {
		return fmt.Errorf("transaction hash mismatch")
	}
	pub, err := phxcrypto.DecodePublicKey(t.PublicKey)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}
	if phxcrypto.AddressFromPublicKey(pub) != t.From {
		return fmt.Errorf("sender address does not match public key")
	}
	if !phxcrypto.Verify(pub, []byte(t.Hash), t.Signature) {
		return fmt.Errorf("invalid transaction signature")
	}
	return nil
}

func SignTransaction(t *Transaction, priv ed25519.PrivateKey) error {
	pub := priv.Public().(ed25519.PublicKey)
	if t.From == "" {
		t.From = phxcrypto.AddressFromPublicKey(pub)
	}
	t.PublicKey = phxcrypto.EncodePublicKey(pub)
	t.RefreshHash()
	t.Signature = phxcrypto.Sign(priv, []byte(t.Hash))
	return nil
}
