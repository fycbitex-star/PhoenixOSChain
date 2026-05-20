package state_test

import (
	"testing"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/state"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func TestBalanceAndNonceValidation(t *testing.T) {
	pub, priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	from := phxcrypto.AddressFromPublicKey(pub)
	db := state.FromAlloc(map[string]uint64{from: 50})

	transaction := tx.New(from, "phxto", 30, 0, 21000, 1)
	if err := tx.SignTransaction(&transaction, priv); err != nil {
		t.Fatal(err)
	}
	if err := db.ApplyTransaction(transaction); err != nil {
		t.Fatal(err)
	}
	if db.BalanceOf(from) != 20 || db.NonceOf(from) != 1 {
		t.Fatal("state transition failed")
	}

	badNonce := tx.New(from, "phxto", 1, 0, 21000, 1)
	if err := tx.SignTransaction(&badNonce, priv); err != nil {
		t.Fatal(err)
	}
	if err := db.ValidateTransaction(badNonce); err == nil {
		t.Fatal("expected nonce rejection")
	}
}

func TestRejectsOverspend(t *testing.T) {
	pub, priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	from := phxcrypto.AddressFromPublicKey(pub)
	db := state.FromAlloc(map[string]uint64{from: 5})
	transaction := tx.New(from, "phxto", 6, 0, 21000, 1)
	if err := tx.SignTransaction(&transaction, priv); err != nil {
		t.Fatal(err)
	}
	if err := db.ValidateTransaction(transaction); err == nil {
		t.Fatal("expected overspend rejection")
	}
}
