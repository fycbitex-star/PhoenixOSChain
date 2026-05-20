package tx_test

import (
	"testing"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func TestTransactionSigningAndVerification(t *testing.T) {
	pub, priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	from := phxcrypto.AddressFromPublicKey(pub)
	transaction := tx.New(from, "phxto0000000000000000000000000000000000", 10, 0, 21000, 1)

	if err := tx.SignTransaction(&transaction, priv); err != nil {
		t.Fatal(err)
	}
	if transaction.Hash == "" || transaction.Signature == "" {
		t.Fatal("expected hash and signature")
	}
	if err := transaction.VerifySignature(); err != nil {
		t.Fatal(err)
	}
}

func TestTransactionSignatureRejectsTampering(t *testing.T) {
	pub, priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	transaction := tx.New(phxcrypto.AddressFromPublicKey(pub), "phxto", 10, 0, 21000, 1)
	if err := tx.SignTransaction(&transaction, priv); err != nil {
		t.Fatal(err)
	}
	transaction.Amount = 11
	if err := transaction.VerifySignature(); err == nil {
		t.Fatal("expected tampered transaction rejection")
	}
}

func TestMempoolDuplicatePrevention(t *testing.T) {
	pool := tx.NewMempool(10, nil)
	transaction := tx.Transaction{Hash: "abc"}
	if err := pool.Add(transaction); err != nil {
		t.Fatal(err)
	}
	if err := pool.Add(transaction); err == nil {
		t.Fatal("expected duplicate rejection")
	}
}
