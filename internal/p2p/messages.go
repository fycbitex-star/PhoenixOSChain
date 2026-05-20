package p2p

import (
	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type TransactionMessage struct {
	FromNode    string         `json:"from_node"`
	Transaction tx.Transaction `json:"transaction"`
}

type BlockMessage struct {
	FromNode string      `json:"from_node"`
	Block    chain.Block `json:"block"`
}

type Status struct {
	NodeID string `json:"node_id"`
	Height uint64 `json:"height"`
	Head   string `json:"head"`
}
