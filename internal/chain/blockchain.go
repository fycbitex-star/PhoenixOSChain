package chain

import (
	"fmt"
	"sort"
	"sync"

	"github.com/phoenixchain/phoenixchain/internal/state"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type Blockchain struct {
	mu            sync.RWMutex
	blocks        []Block
	state         *state.StateDB
	validators    map[string]bool
	validatorList []string
}

func NewBlockchain(genesis Genesis) *Blockchain {
	genesisBlock := Block{
		Height:           0,
		Timestamp:        0,
		PreviousHash:     "",
		Transactions:     []tx.Transaction{},
		StateRoot:        "state-root-placeholder",
		TxRoot:           ComputeTxRoot(nil),
		ValidatorAddress: "genesis",
		ValidatorPubKey:  "",
		Signature:        "",
	}
	genesisBlock.RefreshHash()

	validators := make(map[string]bool)
	for _, validator := range genesis.Consensus.Validators {
		validators[validator] = true
	}
	validatorList := normalizeValidators(genesis.Consensus.Validators)

	return &Blockchain{
		blocks:        []Block{genesisBlock},
		state:         state.FromAlloc(genesis.Alloc),
		validators:    validators,
		validatorList: validatorList,
	}
}

func FromBlocks(genesis Genesis, blocks []Block, accounts map[string]state.Account) *Blockchain {
	validators := make(map[string]bool)
	for _, validator := range genesis.Consensus.Validators {
		validators[validator] = true
	}
	validatorList := normalizeValidators(genesis.Consensus.Validators)
	if len(blocks) == 0 {
		return NewBlockchain(genesis)
	}
	stateDB := state.NewStateDB()
	for _, account := range accounts {
		stateDB.SetAccount(account)
	}
	return &Blockchain{
		blocks:        blocks,
		state:         stateDB,
		validators:    validators,
		validatorList: validatorList,
	}
}

func (bc *Blockchain) Head() Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks[len(bc.blocks)-1]
}

func (bc *Blockchain) State() *state.StateDB {
	return bc.state
}

func (bc *Blockchain) Height() uint64 {
	return bc.Head().Height
}

func (bc *Blockchain) GetBlock(height uint64) (Block, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if height >= uint64(len(bc.blocks)) {
		return Block{}, false
	}
	return bc.blocks[height], true
}

func (bc *Blockchain) Blocks() []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	blocks := make([]Block, len(bc.blocks))
	copy(blocks, bc.blocks)
	return blocks
}

func (bc *Blockchain) RecentBlocks(limit int) []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if limit <= 0 || limit > len(bc.blocks) {
		limit = len(bc.blocks)
	}
	blocks := make([]Block, 0, limit)
	for i := len(bc.blocks) - 1; i >= 0 && len(blocks) < limit; i-- {
		blocks = append(blocks, bc.blocks[i])
	}
	return blocks
}

func (bc *Blockchain) FindTransaction(hash string) (tx.Transaction, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	for _, block := range bc.blocks {
		for _, transaction := range block.Transactions {
			if transaction.Hash == hash {
				return transaction, true
			}
		}
	}
	return tx.Transaction{}, false
}

func (bc *Blockchain) RecentTransactions(limit int) []tx.Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if limit <= 0 {
		limit = 25
	}
	transactions := make([]tx.Transaction, 0, limit)
	for i := len(bc.blocks) - 1; i >= 0 && len(transactions) < limit; i-- {
		for j := len(bc.blocks[i].Transactions) - 1; j >= 0 && len(transactions) < limit; j-- {
			transactions = append(transactions, bc.blocks[i].Transactions[j])
		}
	}
	return transactions
}

func (bc *Blockchain) Validators() map[string]bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	validators := make(map[string]bool, len(bc.validators))
	for validator, ok := range bc.validators {
		validators[validator] = ok
	}
	return validators
}

func (bc *Blockchain) ValidatorList() []string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	list := make([]string, len(bc.validatorList))
	copy(list, bc.validatorList)
	return list
}

func (bc *Blockchain) AddValidator(address string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if address == "" {
		return
	}
	if !bc.validators[address] {
		bc.validatorList = append(bc.validatorList, address)
		sort.Strings(bc.validatorList)
	}
	bc.validators[address] = true
}

func (bc *Blockchain) HasBlock(hash string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	for _, block := range bc.blocks {
		if block.Hash == hash {
			return true
		}
	}
	return false
}

func (bc *Blockchain) BuildBlock(transactions []tx.Transaction, validator string) Block {
	head := bc.Head()
	return NewBlock(head.Height+1, head.Hash, transactions, validator)
}

func (bc *Blockchain) AddBlock(block Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	parent := bc.blocks[len(bc.blocks)-1]
	if err := ValidateBlock(block, parent, bc.state, bc.validators, bc.validatorList); err != nil {
		return err
	}
	bc.blocks = append(bc.blocks, block)
	return nil
}

func (bc *Blockchain) ValidateTransaction(transaction tx.Transaction) error {
	if transaction.Hash == "" {
		return fmt.Errorf("transaction hash is required")
	}
	return bc.state.ValidateTransaction(transaction)
}

func (bc *Blockchain) IsProposer(address string, height uint64) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return expectedProposer(address, height, bc.validatorList)
}

func normalizeValidators(validators []string) []string {
	seen := make(map[string]bool)
	list := make([]string, 0, len(validators))
	for _, validator := range validators {
		if validator == "" || seen[validator] {
			continue
		}
		seen[validator] = true
		list = append(list, validator)
	}
	sort.Strings(list)
	return list
}

func expectedProposer(address string, height uint64, validators []string) bool {
	if len(validators) == 0 {
		return false
	}
	return validators[int(height%uint64(len(validators)))] == address
}
