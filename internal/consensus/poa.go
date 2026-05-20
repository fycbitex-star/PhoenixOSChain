package consensus

import (
	"fmt"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type PoA struct {
	cfg Config
}

func NewPoA(cfg Config) *PoA {
	return &PoA{cfg: cfg}
}

func (p *PoA) BuildAndSign(parent chain.Block, transactions []tx.Transaction) (chain.Block, error) {
	if p.cfg.ValidatorKey == nil {
		return chain.Block{}, fmt.Errorf("validator key is required")
	}
	if !p.cfg.Validators[p.cfg.ValidatorAddress] {
		return chain.Block{}, fmt.Errorf("validator is not authorized")
	}
	if len(p.cfg.ValidatorList) > 0 {
		expected := p.cfg.ValidatorList[int((parent.Height+1)%uint64(len(p.cfg.ValidatorList)))]
		if expected != p.cfg.ValidatorAddress {
			return chain.Block{}, fmt.Errorf("validator is not proposer for height %d", parent.Height+1)
		}
	}
	block := chain.NewBlock(parent.Height+1, parent.Hash, transactions, p.cfg.ValidatorAddress)
	if err := block.Sign(p.cfg.ValidatorKey); err != nil {
		return chain.Block{}, err
	}
	return block, nil
}

func (p *PoA) Validate(block chain.Block) error {
	if !p.cfg.Validators[block.ValidatorAddress] {
		return fmt.Errorf("validator is not authorized")
	}
	return block.VerifySignature()
}
