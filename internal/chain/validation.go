package chain

import (
	"fmt"

	"github.com/phoenixchain/phoenixchain/internal/state"
)

func ValidateBlock(block Block, parent Block, stateDB *state.StateDB, validators map[string]bool, validatorList []string) error {
	if block.Height != parent.Height+1 {
		return fmt.Errorf("invalid block height")
	}
	if block.PreviousHash != parent.Hash {
		return fmt.Errorf("invalid previous hash")
	}
	if block.TxRoot != ComputeTxRoot(block.Transactions) {
		return fmt.Errorf("invalid transaction root")
	}
	if block.Hash != block.ComputeHash() {
		return fmt.Errorf("invalid block hash")
	}
	if !validators[block.ValidatorAddress] {
		return fmt.Errorf("unauthorized validator")
	}
	if !expectedProposer(block.ValidatorAddress, block.Height, validatorList) {
		return fmt.Errorf("invalid block proposer")
	}
	if err := block.VerifySignature(); err != nil {
		return err
	}

	seen := make(map[string]bool)
	workingState := state.FromAccounts(stateDB.Snapshot())
	for _, transaction := range block.Transactions {
		if seen[transaction.Hash] {
			return fmt.Errorf("duplicate transaction in block")
		}
		seen[transaction.Hash] = true
		if err := workingState.ApplyTransaction(transaction); err != nil {
			return err
		}
	}
	for _, account := range workingState.Snapshot() {
		stateDB.SetAccount(account)
	}
	return nil
}
