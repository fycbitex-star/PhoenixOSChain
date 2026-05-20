package state

import (
	"fmt"
	"sync"

	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type StateDB struct {
	mu       sync.RWMutex
	accounts map[string]Account
}

func NewStateDB() *StateDB {
	return &StateDB{accounts: make(map[string]Account)}
}

func FromAlloc(alloc map[string]uint64) *StateDB {
	db := NewStateDB()
	for address, balance := range alloc {
		db.accounts[address] = Account{Address: address, Balance: balance}
	}
	return db
}

func FromAccounts(accounts map[string]Account) *StateDB {
	db := NewStateDB()
	for address, account := range accounts {
		db.accounts[address] = account
	}
	return db
}

func (db *StateDB) GetAccount(address string) Account {
	db.mu.RLock()
	defer db.mu.RUnlock()
	account, ok := db.accounts[address]
	if !ok {
		return Account{Address: address}
	}
	return account
}

func (db *StateDB) SetAccount(account Account) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.accounts[account.Address] = account
}

func (db *StateDB) Snapshot() map[string]Account {
	db.mu.RLock()
	defer db.mu.RUnlock()
	snapshot := make(map[string]Account, len(db.accounts))
	for address, account := range db.accounts {
		snapshot[address] = account
	}
	return snapshot
}

func (db *StateDB) ValidateTransaction(transaction tx.Transaction) error {
	if err := transaction.VerifySignature(); err != nil {
		return err
	}
	sender := db.GetAccount(transaction.From)
	if sender.Balance < transaction.Amount {
		return fmt.Errorf("insufficient balance")
	}
	if sender.Nonce != transaction.Nonce {
		return fmt.Errorf("invalid nonce: expected %d got %d", sender.Nonce, transaction.Nonce)
	}
	return nil
}

func (db *StateDB) ApplyTransaction(transaction tx.Transaction) error {
	if err := db.ValidateTransaction(transaction); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()
	sender := db.accounts[transaction.From]
	recipient := db.accounts[transaction.To]
	recipient.Address = transaction.To

	sender.Balance -= transaction.Amount
	sender.Nonce++
	recipient.Balance += transaction.Amount

	db.accounts[transaction.From] = sender
	db.accounts[transaction.To] = recipient
	return nil
}

func (db *StateDB) RollbackPlaceholder() error {
	return fmt.Errorf("rollback is not implemented in v0.1")
}
