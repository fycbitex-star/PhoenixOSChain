package wallet_test

import (
	"path/filepath"
	"testing"

	"github.com/phoenixchain/phoenixchain/internal/tx"
	"github.com/phoenixchain/phoenixchain/internal/wallet"
)

func TestEncryptedKeystoreRoundTripAndSigning(t *testing.T) {
	w, err := wallet.New()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "wallet.json")
	if err := wallet.SaveEncrypted(path, w, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	loaded, err := wallet.LoadEncrypted(path, "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Address != w.Address {
		t.Fatalf("address = %s, want %s", loaded.Address, w.Address)
	}
	transaction := tx.New(loaded.Address, "phxto", 10, 0, 21000, 1)
	if err := loaded.SignTransaction(&transaction); err != nil {
		t.Fatal(err)
	}
	if err := transaction.VerifySignature(); err != nil {
		t.Fatal(err)
	}
}

func TestEncryptedKeystoreRejectsWrongPassword(t *testing.T) {
	w, err := wallet.New()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "wallet.json")
	if err := wallet.SaveEncrypted(path, w, "right"); err != nil {
		t.Fatal(err)
	}
	if _, err := wallet.LoadEncrypted(path, "wrong"); err == nil {
		t.Fatal("expected wrong password rejection")
	}
}

func TestMnemonicGeneration(t *testing.T) {
	mnemonic, err := wallet.GenerateMnemonic()
	if err != nil {
		t.Fatal(err)
	}
	if len(mnemonic) != 32 {
		t.Fatalf("mnemonic length = %d, want 32", len(mnemonic))
	}
}
