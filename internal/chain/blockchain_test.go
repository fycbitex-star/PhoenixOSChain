package chain_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func TestBlockValidationAndStateUpdate(t *testing.T) {
	pub, priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	validatorPub, validatorPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}

	from := phxcrypto.AddressFromPublicKey(pub)
	to := "phxrecipient000000000000000000000000000000"
	validator := phxcrypto.AddressFromPublicKey(validatorPub)
	genesis := chain.DefaultGenesis()
	genesis.Alloc[from] = 100
	genesis.Consensus.Validators = []string{validator}

	bc := chain.NewBlockchain(genesis)
	transaction := tx.New(from, to, 25, 0, 21000, 1)
	if err := tx.SignTransaction(&transaction, priv); err != nil {
		t.Fatal(err)
	}

	block := bc.BuildBlock([]tx.Transaction{transaction}, validator)
	if err := block.Sign(validatorPriv); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddBlock(block); err != nil {
		t.Fatal(err)
	}

	if got := bc.State().BalanceOf(from); got != 75 {
		t.Fatalf("sender balance = %d, want 75", got)
	}
	if got := bc.State().NonceOf(from); got != 1 {
		t.Fatalf("sender nonce = %d, want 1", got)
	}
	if got := bc.State().BalanceOf(to); got != 25 {
		t.Fatalf("recipient balance = %d, want 25", got)
	}
}

func TestRejectsInvalidBlockPreviousHash(t *testing.T) {
	_, validatorPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	validator := phxcrypto.AddressFromPublicKey(validatorPriv.Public().(ed25519.PublicKey))
	genesis := chain.DefaultGenesis()
	genesis.Consensus.Validators = []string{validator}
	bc := chain.NewBlockchain(genesis)

	block := chain.NewBlock(1, "bad-parent", nil, validator)
	if err := block.Sign(validatorPriv); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddBlock(block); err == nil {
		t.Fatal("expected invalid previous hash rejection")
	}
}
