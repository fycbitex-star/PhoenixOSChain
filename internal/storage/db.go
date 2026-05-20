package storage

import (
	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/governance"
	"github.com/phoenixchain/phoenixchain/internal/phx20"
	"github.com/phoenixchain/phoenixchain/internal/state"
)

type Snapshot struct {
	Genesis    chain.Genesis            `json:"genesis"`
	Blocks     []chain.Block            `json:"blocks"`
	Accounts   map[string]state.Account `json:"accounts"`
	PHX20      phx20.Snapshot           `json:"phx20,omitempty"`
	Governance governance.Snapshot      `json:"governance,omitempty"`
}

type DB interface {
	Load() (Snapshot, error)
	Save(Snapshot) error
}
