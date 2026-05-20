package chain

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type Block struct {
	Height           uint64           `json:"height"`
	Timestamp        int64            `json:"timestamp"`
	PreviousHash     string           `json:"previous_hash"`
	Transactions     []tx.Transaction `json:"transactions"`
	StateRoot        string           `json:"state_root"`
	TxRoot           string           `json:"tx_root"`
	ValidatorAddress string           `json:"validator_address"`
	ValidatorPubKey  string           `json:"validator_pub_key"`
	Signature        string           `json:"signature"`
	Hash             string           `json:"hash"`
}

type unsignedBlock struct {
	Height           uint64           `json:"height"`
	Timestamp        int64            `json:"timestamp"`
	PreviousHash     string           `json:"previous_hash"`
	Transactions     []tx.Transaction `json:"transactions"`
	StateRoot        string           `json:"state_root"`
	TxRoot           string           `json:"tx_root"`
	ValidatorAddress string           `json:"validator_address"`
	ValidatorPubKey  string           `json:"validator_pub_key"`
}

func NewBlock(height uint64, previousHash string, transactions []tx.Transaction, validatorAddress string) Block {
	block := Block{
		Height:           height,
		Timestamp:        time.Now().UTC().Unix(),
		PreviousHash:     previousHash,
		Transactions:     transactions,
		StateRoot:        "state-root-placeholder",
		ValidatorAddress: validatorAddress,
	}
	block.TxRoot = ComputeTxRoot(transactions)
	block.RefreshHash()
	return block
}

func (b Block) unsigned() unsignedBlock {
	return unsignedBlock{
		Height:           b.Height,
		Timestamp:        b.Timestamp,
		PreviousHash:     b.PreviousHash,
		Transactions:     b.Transactions,
		StateRoot:        b.StateRoot,
		TxRoot:           b.TxRoot,
		ValidatorAddress: b.ValidatorAddress,
		ValidatorPubKey:  b.ValidatorPubKey,
	}
}

func (b Block) ComputeHash() string {
	payload, _ := json.Marshal(b.unsigned())
	return phxcrypto.HashBytes(payload)
}

func (b *Block) RefreshHash() {
	b.TxRoot = ComputeTxRoot(b.Transactions)
	b.Hash = b.ComputeHash()
}

func (b *Block) Sign(priv ed25519.PrivateKey) error {
	pub := priv.Public().(ed25519.PublicKey)
	address := phxcrypto.AddressFromPublicKey(pub)
	if b.ValidatorAddress == "" {
		b.ValidatorAddress = address
	}
	if b.ValidatorAddress != address {
		return fmt.Errorf("validator private key does not match block validator")
	}
	b.ValidatorPubKey = phxcrypto.EncodePublicKey(pub)
	b.RefreshHash()
	b.Signature = phxcrypto.Sign(priv, []byte(b.Hash))
	return nil
}

func (b Block) VerifySignature() error {
	pub, err := phxcrypto.DecodePublicKey(b.ValidatorPubKey)
	if err != nil {
		return err
	}
	if phxcrypto.AddressFromPublicKey(pub) != b.ValidatorAddress {
		return fmt.Errorf("validator address does not match public key")
	}
	if !phxcrypto.Verify(pub, []byte(b.Hash), b.Signature) {
		return fmt.Errorf("invalid block signature")
	}
	return nil
}

func ComputeTxRoot(transactions []tx.Transaction) string {
	hashes := make([]string, 0, len(transactions))
	for _, transaction := range transactions {
		hashes = append(hashes, transaction.Hash)
	}
	root, _ := phxcrypto.HashJSON(hashes)
	return root
}
