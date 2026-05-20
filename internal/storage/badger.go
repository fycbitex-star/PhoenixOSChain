package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	badger "github.com/dgraph-io/badger/v4"
)

const snapshotKey = "phoenixchain:snapshot:v1"
const metaKey = "phoenixchain:meta:v1"

type BadgerDB struct {
	path string
}

func NewBadgerDB(dataDir string) *BadgerDB {
	return &BadgerDB{path: filepath.Join(dataDir, "badger")}
}

func (db *BadgerDB) Load() (Snapshot, error) {
	handle, err := db.open()
	if err != nil {
		return Snapshot{}, err
	}
	defer handle.Close()

	var snapshot Snapshot
	err = handle.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(snapshotKey))
		if err != nil {
			return err
		}
		return item.Value(func(value []byte) error {
			return json.Unmarshal(value, &snapshot)
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return Snapshot{}, os.ErrNotExist
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("load badger snapshot: %w", err)
	}
	if err := validateSnapshot(snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (db *BadgerDB) Save(snapshot Snapshot) error {
	handle, err := db.open()
	if err != nil {
		return err
	}
	defer handle.Close()

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return handle.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(snapshotKey), data); err != nil {
			return err
		}
		meta, _ := json.Marshal(map[string]any{
			"height":   len(snapshot.Blocks) - 1,
			"accounts": len(snapshot.Accounts),
		})
		if err := txn.Set([]byte(metaKey), meta); err != nil {
			return err
		}
		for _, block := range snapshot.Blocks {
			blockData, err := json.Marshal(block)
			if err != nil {
				return err
			}
			if err := txn.Set([]byte(fmt.Sprintf("block:height:%020d", block.Height)), blockData); err != nil {
				return err
			}
			if err := txn.Set([]byte("block:hash:"+block.Hash), blockData); err != nil {
				return err
			}
			for _, transaction := range block.Transactions {
				txData, err := json.Marshal(transaction)
				if err != nil {
					return err
				}
				if err := txn.Set([]byte("tx:"+transaction.Hash), txData); err != nil {
					return err
				}
			}
		}
		for address, account := range snapshot.Accounts {
			accountData, err := json.Marshal(account)
			if err != nil {
				return err
			}
			if err := txn.Set([]byte("account:"+address), accountData); err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *BadgerDB) open() (*badger.DB, error) {
	if err := os.MkdirAll(db.path, 0o755); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(db.path)
	opts.Logger = nil
	return badger.Open(opts)
}

func validateSnapshot(snapshot Snapshot) error {
	if len(snapshot.Blocks) == 0 {
		return fmt.Errorf("badger snapshot has no blocks")
	}
	for i, block := range snapshot.Blocks {
		if block.Height != uint64(i) {
			return fmt.Errorf("badger snapshot height mismatch at %d", i)
		}
		if block.Hash != block.ComputeHash() {
			return fmt.Errorf("badger snapshot block hash mismatch at height %d", block.Height)
		}
		if i > 0 && block.PreviousHash != snapshot.Blocks[i-1].Hash {
			return fmt.Errorf("badger snapshot parent mismatch at height %d", block.Height)
		}
	}
	return nil
}
