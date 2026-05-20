package tx

import (
	"fmt"
	"sort"
	"sync"
)

type ValidatorFunc func(Transaction) error

type Mempool struct {
	mu       sync.RWMutex
	maxSize  int
	byHash   map[string]Transaction
	validate ValidatorFunc
}

func NewMempool(maxSize int, validate ValidatorFunc) *Mempool {
	return &Mempool{
		maxSize:  maxSize,
		byHash:   make(map[string]Transaction),
		validate: validate,
	}
}

func (m *Mempool) Add(transaction Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.byHash) >= m.maxSize {
		return fmt.Errorf("mempool is full")
	}
	if _, exists := m.byHash[transaction.Hash]; exists {
		return fmt.Errorf("duplicate transaction")
	}
	if m.validate != nil {
		if err := m.validate(transaction); err != nil {
			return err
		}
	}
	m.byHash[transaction.Hash] = transaction
	return nil
}

func (m *Mempool) Get(hash string) (Transaction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	transaction, ok := m.byHash[hash]
	return transaction, ok
}

func (m *Mempool) All() []Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transactions := make([]Transaction, 0, len(m.byHash))
	for _, transaction := range m.byHash {
		transactions = append(transactions, transaction)
	}
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].GasPrice == transactions[j].GasPrice {
			return transactions[i].Hash < transactions[j].Hash
		}
		return transactions[i].GasPrice > transactions[j].GasPrice
	})
	return transactions
}

func (m *Mempool) RemoveConfirmed(transactions []Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, transaction := range transactions {
		delete(m.byHash, transaction.Hash)
	}
}
