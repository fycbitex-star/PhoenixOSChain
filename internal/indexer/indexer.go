package indexer

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/state"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

const (
	DefaultPageSize  = 25
	MaxPageSize      = 100
	ContractPrefix   = "pc"
	AddressPrefix    = "phx"
	IndexerStaleTime = 120 * time.Second
)

type Health string

const (
	HealthReady   Health = "ready"
	HealthLagging Health = "lagging"
	HealthStale   Health = "stale"
	HealthDegraded Health = "degraded"
)

type Status struct {
	IndexedHeight     uint64    `json:"indexed_height"`
	ChainHeadHeight   uint64    `json:"chain_head_height"`
	IndexingLag       uint64    `json:"indexing_lag"`
	Sequential        bool      `json:"sequential"`
	MissingBlocks     []uint64  `json:"missing_blocks"`
	LastIndexedAt     time.Time `json:"last_indexed_at"`
	LastIndexedReason string    `json:"last_indexed_reason"`
	Health            Health    `json:"health"`
	BlockCount        int       `json:"block_count"`
	TxCount           int       `json:"tx_count"`
	AddressCount      int       `json:"address_count"`
	ContractCount     int       `json:"contract_count"`
	TokenTransfers    int       `json:"token_transfers"`
}

type BlockRecord struct {
	Height           uint64           `json:"height"`
	Hash             string           `json:"hash"`
	PreviousHash     string           `json:"previous_hash"`
	Timestamp        int64            `json:"timestamp"`
	TimestampLabel   string           `json:"timestamp_label"`
	ValidatorAddress string           `json:"validator_address"`
	StateRoot        string           `json:"state_root"`
	TxRoot           string           `json:"tx_root"`
	TransactionCount int              `json:"transaction_count"`
	Transactions     []tx.Transaction `json:"transactions"`
}

type TransactionRecord struct {
	Hash             string   `json:"hash"`
	BlockHeight      uint64   `json:"block_height"`
	BlockHash        string   `json:"block_hash"`
	Timestamp        int64    `json:"timestamp"`
	TimestampLabel   string   `json:"timestamp_label"`
	Status           string   `json:"status"`
	From             string   `json:"from"`
	To               string   `json:"to"`
	Amount           uint64   `json:"amount"`
	GasLimit         uint64   `json:"gas_limit"`
	GasPrice         uint64   `json:"gas_price"`
	Nonce            uint64   `json:"nonce"`
	Type             string   `json:"type"`
	Method           string   `json:"method"`
	InputAvailable   bool     `json:"input_available"`
	ReceiptAvailable bool     `json:"receipt_available"`
	LogsAvailable    bool     `json:"logs_available"`
	RelatedAddresses []string `json:"related_addresses"`
}

type ActivityCounter struct {
	Transactions        int `json:"transactions"`
	TransfersSent       int `json:"transfers_sent"`
	TransfersReceived   int `json:"transfers_received"`
	ContractCalls       int `json:"contract_calls"`
	ContractDeployments int `json:"contract_deployments"`
	ValidatorActions    int `json:"validator_actions"`
	StakingActions      int `json:"staking_actions"`
	TokenTransfers      int `json:"token_transfers"`
}

type AddressRecord struct {
	Address         string              `json:"address"`
	Balance         uint64              `json:"balance"`
	Nonce           uint64              `json:"nonce"`
	RecentActivity  []TransactionRecord `json:"recent_activity"`
	ActivityCounter ActivityCounter     `json:"activity_counter"`
}

type ContractRecord struct {
	Address            string              `json:"address"`
	RuntimeStatus      string              `json:"runtime_status"`
	ContractType       string              `json:"contract_type"`
	Creator            string              `json:"creator"`
	CreationTxHash     string              `json:"creation_tx_hash"`
	VerifiedSource     bool                `json:"verified_source"`
	ABIAvailable       bool                `json:"abi_available"`
	InteractionCount   int                 `json:"interaction_count"`
	RecentInteractions []TransactionRecord `json:"recent_interactions"`
	RecentEvents       []string            `json:"recent_events"`
}

type TokenIndexState struct {
	Supported            bool   `json:"supported"`
	IndexedTransferCount int    `json:"indexed_transfer_count"`
	Status               string `json:"status"`
	Note                 string `json:"note"`
}

type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type RuntimeIndexer struct {
	mu            sync.RWMutex
	status        Status
	blocks        []BlockRecord
	blockByHash   map[string]BlockRecord
	txs           []TransactionRecord
	txByHash      map[string]TransactionRecord
	addresses     map[string]AddressRecord
	contracts     map[string]ContractRecord
	tokenState    TokenIndexState
}

func New() *RuntimeIndexer {
	return &RuntimeIndexer{
		blockByHash: make(map[string]BlockRecord),
		txByHash:    make(map[string]TransactionRecord),
		addresses:   make(map[string]AddressRecord),
		contracts:   make(map[string]ContractRecord),
		tokenState: TokenIndexState{
			Supported: false,
			Status:    "prototype",
			Note:      "Token transfer indexing is pending runtime support and event-backed token standards.",
		},
	}
}

func (ri *RuntimeIndexer) Rebuild(blocks []chain.Block, accounts map[string]state.Account, reason string) {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	now := time.Now().UTC()
	blockRecords := make([]BlockRecord, 0, len(blocks))
	blockByHash := make(map[string]BlockRecord, len(blocks))
	txRecords := make([]TransactionRecord, 0)
	txByHash := make(map[string]TransactionRecord)
	addresses := make(map[string]AddressRecord, len(accounts))
	contracts := make(map[string]ContractRecord)
	missingBlocks := make([]uint64, 0)
	sequential := true
	var lastHeight uint64

	for address, account := range accounts {
		addresses[address] = AddressRecord{
			Address: address,
			Balance: account.Balance,
			Nonce:   account.Nonce,
		}
	}

	for i, block := range blocks {
		if i > 0 && block.Height != lastHeight+1 {
			sequential = false
			for gap := lastHeight + 1; gap < block.Height; gap++ {
				missingBlocks = append(missingBlocks, gap)
			}
		}
		lastHeight = block.Height

		record := BlockRecord{
			Height:           block.Height,
			Hash:             block.Hash,
			PreviousHash:     block.PreviousHash,
			Timestamp:        block.Timestamp,
			TimestampLabel:   formatTimestamp(block.Timestamp),
			ValidatorAddress: block.ValidatorAddress,
			StateRoot:        block.StateRoot,
			TxRoot:           block.TxRoot,
			TransactionCount: len(block.Transactions),
			Transactions:     append([]tx.Transaction(nil), block.Transactions...),
		}
		blockRecords = append(blockRecords, record)
		blockByHash[record.Hash] = record

		for _, item := range block.Transactions {
			txRecord := TransactionRecord{
				Hash:             item.Hash,
				BlockHeight:      block.Height,
				BlockHash:        block.Hash,
				Timestamp:        block.Timestamp,
				TimestampLabel:   formatTimestamp(block.Timestamp),
				Status:           "confirmed",
				From:             item.From,
				To:               item.To,
				Amount:           item.Amount,
				GasLimit:         item.GasLimit,
				GasPrice:         item.GasPrice,
				Nonce:            item.Nonce,
				Type:             classifyTransaction(item),
				Method:           classifyMethod(item),
				InputAvailable:   false,
				ReceiptAvailable: false,
				LogsAvailable:    false,
				RelatedAddresses: relatedAddresses(item),
			}
			txRecords = append(txRecords, txRecord)
			txByHash[txRecord.Hash] = txRecord
			indexAddressActivity(addresses, txRecord)
			indexContractActivity(contracts, txRecord)
		}
	}

	sort.Slice(blockRecords, func(i, j int) bool { return blockRecords[i].Height > blockRecords[j].Height })
	sort.Slice(txRecords, func(i, j int) bool {
		if txRecords[i].BlockHeight == txRecords[j].BlockHeight {
			return txRecords[i].Hash > txRecords[j].Hash
		}
		return txRecords[i].BlockHeight > txRecords[j].BlockHeight
	})

	for address, record := range addresses {
		sort.Slice(record.RecentActivity, func(i, j int) bool {
			if record.RecentActivity[i].BlockHeight == record.RecentActivity[j].BlockHeight {
				return record.RecentActivity[i].Hash > record.RecentActivity[j].Hash
			}
			return record.RecentActivity[i].BlockHeight > record.RecentActivity[j].BlockHeight
		})
		if len(record.RecentActivity) > 100 {
			record.RecentActivity = record.RecentActivity[:100]
		}
		addresses[address] = record
	}
	for address, record := range contracts {
		sort.Slice(record.RecentInteractions, func(i, j int) bool {
			if record.RecentInteractions[i].BlockHeight == record.RecentInteractions[j].BlockHeight {
				return record.RecentInteractions[i].Hash > record.RecentInteractions[j].Hash
			}
			return record.RecentInteractions[i].BlockHeight > record.RecentInteractions[j].BlockHeight
		})
		if len(record.RecentInteractions) > 50 {
			record.RecentInteractions = record.RecentInteractions[:50]
		}
		contracts[address] = record
	}

	chainHead := uint64(0)
	if len(blocks) > 0 {
		chainHead = blocks[len(blocks)-1].Height
	}
	health := HealthReady
	if !sequential {
		health = HealthDegraded
	}

	ri.blocks = blockRecords
	ri.blockByHash = blockByHash
	ri.txs = txRecords
	ri.txByHash = txByHash
	ri.addresses = addresses
	ri.contracts = contracts
	ri.status = Status{
		IndexedHeight:     chainHead,
		ChainHeadHeight:   chainHead,
		IndexingLag:       0,
		Sequential:        sequential,
		MissingBlocks:     missingBlocks,
		LastIndexedAt:     now,
		LastIndexedReason: reason,
		Health:            health,
		BlockCount:        len(blockRecords),
		TxCount:           len(txRecords),
		AddressCount:      len(addresses),
		ContractCount:     len(contracts),
		TokenTransfers:    0,
	}
}

func (ri *RuntimeIndexer) Status() Status {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	status := ri.status
	if !status.LastIndexedAt.IsZero() && time.Since(status.LastIndexedAt) > IndexerStaleTime && status.Health == HealthReady {
		status.Health = HealthStale
	}
	return status
}

func (ri *RuntimeIndexer) RecentBlocks(page, pageSize int) Page[BlockRecord] {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	return paginate(ri.blocks, page, pageSize)
}

func (ri *RuntimeIndexer) RecentTransactions(page, pageSize int) Page[TransactionRecord] {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	return paginate(ri.txs, page, pageSize)
}

func (ri *RuntimeIndexer) AddressActivity(address string, page, pageSize int) (AddressRecord, Page[TransactionRecord], bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	record, ok := ri.addresses[address]
	if !ok {
		return AddressRecord{}, Page[TransactionRecord]{}, false
	}
	return record, paginate(record.RecentActivity, page, pageSize), true
}

func (ri *RuntimeIndexer) FindBlockByHash(hash string) (BlockRecord, bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	record, ok := ri.blockByHash[hash]
	return record, ok
}

func (ri *RuntimeIndexer) FindBlockByHeight(height uint64) (BlockRecord, bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	for _, record := range ri.blocks {
		if record.Height == height {
			return record, true
		}
	}
	return BlockRecord{}, false
}

func (ri *RuntimeIndexer) FindTransaction(hash string) (TransactionRecord, bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	record, ok := ri.txByHash[hash]
	return record, ok
}

func (ri *RuntimeIndexer) FindContract(address string) (ContractRecord, bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	record, ok := ri.contracts[address]
	return record, ok
}

func (ri *RuntimeIndexer) ContractsPage(page, pageSize int) Page[ContractRecord] {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	records := make([]ContractRecord, 0, len(ri.contracts))
	for _, record := range ri.contracts {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].InteractionCount > records[j].InteractionCount })
	return paginate(records, page, pageSize)
}

func (ri *RuntimeIndexer) TokenState() TokenIndexState {
	ri.mu.RLock()
	defer ri.mu.RUnlock()
	return ri.tokenState
}

func classifyTransaction(item tx.Transaction) string {
	if strings.HasPrefix(item.To, ContractPrefix) {
		return "call_contract"
	}
	return "transfer"
}

func classifyMethod(item tx.Transaction) string {
	if strings.HasPrefix(item.To, ContractPrefix) {
		return "contract_call"
	}
	return "native_transfer"
}

func relatedAddresses(item tx.Transaction) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 2)
	for _, value := range []string{item.From, item.To} {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func indexAddressActivity(addresses map[string]AddressRecord, txRecord TransactionRecord) {
	for _, address := range txRecord.RelatedAddresses {
		record := addresses[address]
		record.Address = address
		record.RecentActivity = append(record.RecentActivity, txRecord)
		record.ActivityCounter.Transactions++
		if txRecord.Type == "call_contract" {
			record.ActivityCounter.ContractCalls++
		}
		if txRecord.From == address {
			record.ActivityCounter.TransfersSent++
		}
		if txRecord.To == address {
			record.ActivityCounter.TransfersReceived++
		}
		addresses[address] = record
	}
}

func indexContractActivity(contracts map[string]ContractRecord, txRecord TransactionRecord) {
	if !strings.HasPrefix(txRecord.To, ContractPrefix) {
		return
	}
	record := contracts[txRecord.To]
	record.Address = txRecord.To
	record.RuntimeStatus = "prototype runtime"
	record.ContractType = "pcvm-compatible placeholder"
	record.InteractionCount++
	record.RecentInteractions = append(record.RecentInteractions, txRecord)
	contracts[txRecord.To] = record
}

func formatTimestamp(value int64) string {
	if value <= 0 {
		return "genesis"
	}
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}

func paginate[T any](items []T, page, pageSize int) Page[T] {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if page <= 0 {
		page = 1
	}
	totalItems := len(items)
	totalPages := totalItems / pageSize
	if totalItems%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	start := (page - 1) * pageSize
	if start > totalItems {
		start = totalItems
	}
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}
	out := make([]T, end-start)
	copy(out, items[start:end])
	return Page[T]{
		Items:      out,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func NormalizeSearchQuery(value string) (string, string) {
	query := strings.TrimSpace(strings.ToLower(value))
	switch {
	case query == "":
		return "empty", ""
	case strings.HasPrefix(query, AddressPrefix) && len(query) == 43:
		return "address", query
	case strings.HasPrefix(query, ContractPrefix) && len(query) == 42:
		return "contract", query
	case isDigits(query):
		return "block_height", query
	case isHex(query) && len(query) >= 40:
		return "hash", query
	default:
		return "unknown", query
	}
}

func isDigits(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return value != ""
}

func isHex(value string) bool {
	for _, ch := range value {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return value != ""
}

func (s Status) Summary() string {
	return fmt.Sprintf("indexed=%d lag=%d health=%s", s.IndexedHeight, s.IndexingLag, s.Health)
}
