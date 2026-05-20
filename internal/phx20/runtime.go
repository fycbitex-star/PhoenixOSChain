package phx20

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/vm"
)

const (
	StandardName    = "PHX-20"
	ZeroAddress     = "phx0000000000000000000000000000000000000000"
	MaxNameLength   = 48
	MaxSymbolLength = 12
	MaxDecimals     = 18
	MaxURLLength    = 160
	MaxDescriptionLength = 280
)

type Spec struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Runtime        string   `json:"runtime"`
	StandardBadge  string   `json:"standard_badge"`
	RequiredFields []string `json:"required_fields"`
	RequiredEvents []string `json:"required_events"`
	OptionalCaps   []string `json:"optional_caps"`
	IntegrityNotes []string `json:"integrity_notes"`
}

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mintable    bool   `json:"mintable"`
	Burnable    bool   `json:"burnable"`
	Pausable    bool   `json:"pausable"`
	Governance  bool   `json:"governance"`
	GasEstimate uint64 `json:"gas_estimate"`
}

type DeployConfig struct {
	Creator   string `json:"creator"`
	Owner     string `json:"owner"`
	Treasury  string `json:"treasury"`
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Decimals  uint8  `json:"decimals"`
	Supply    uint64 `json:"supply"`
	Template  string `json:"template"`
	Mintable  bool   `json:"mintable"`
	Burnable  bool   `json:"burnable"`
	Pausable  bool   `json:"pausable"`
	LogoURL      string `json:"logo_url"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url"`
	TwitterURL   string `json:"twitter_url"`
	GithubURL    string `json:"github_url"`
}

type Event struct {
	ID           string `json:"id"`
	TokenAddress string `json:"token_address"`
	Type         string `json:"type"`
	TxHash       string `json:"tx_hash"`
	From         string `json:"from,omitempty"`
	To           string `json:"to,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Spender      string `json:"spender,omitempty"`
	Amount       uint64 `json:"amount,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	Label        string `json:"label"`
}

type Token struct {
	Address         string           `json:"address"`
	Creator         string           `json:"creator"`
	Owner           string           `json:"owner"`
	Treasury        string           `json:"treasury"`
	Name            string           `json:"name"`
	Symbol          string           `json:"symbol"`
	Decimals        uint8            `json:"decimals"`
	TotalSupply     uint64           `json:"total_supply"`
	DeployTxHash    string           `json:"deploy_tx_hash"`
	DeployNonce     uint64           `json:"deploy_nonce"`
	Verified        bool             `json:"verified"`
	CreatedAt       int64            `json:"created_at"`
	StandardBadge   string           `json:"standard_badge"`
	RuntimeStatus   string           `json:"runtime_status"`
	AuditStatus     string           `json:"audit_status"`
	Template        string           `json:"template"`
	LogoURL         string           `json:"logo_url"`
	Description     string           `json:"description"`
	WebsiteURL      string           `json:"website_url"`
	TwitterURL      string           `json:"twitter_url"`
	GithubURL       string           `json:"github_url"`
	Mintable        bool             `json:"mintable"`
	Burnable        bool             `json:"burnable"`
	Pausable        bool             `json:"pausable"`
	Paused          bool             `json:"paused"`
	DeployGasUsed   uint64           `json:"deploy_gas_used"`
	DeployLogCount  int              `json:"deploy_log_count"`
	InteractionCount int             `json:"interaction_count"`
	Holders         map[string]uint64 `json:"holders"`
}

type TokenSummary struct {
	Address       string `json:"address"`
	Creator       string `json:"creator"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	Decimals      uint8  `json:"decimals"`
	TotalSupply   uint64 `json:"total_supply"`
	DeployTxHash  string `json:"deploy_tx_hash"`
	Verified      bool   `json:"verified"`
	CreatedAt     int64  `json:"created_at"`
	StandardBadge string `json:"standard_badge"`
	RuntimeStatus string `json:"runtime_status"`
	AuditStatus   string `json:"audit_status"`
	Template      string `json:"template"`
	LogoURL       string `json:"logo_url"`
	Description   string `json:"description"`
	WebsiteURL    string `json:"website_url"`
	TwitterURL    string `json:"twitter_url"`
	GithubURL     string `json:"github_url"`
	Mintable      bool   `json:"mintable"`
	Burnable      bool   `json:"burnable"`
	Pausable      bool   `json:"pausable"`
	Paused        bool   `json:"paused"`
	HoldersCount  int    `json:"holders_count"`
	TransferCount int    `json:"transfer_count"`
}

type HolderBalance struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

type WalletBalance struct {
	TokenAddress string `json:"token_address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     uint8  `json:"decimals"`
	Balance      uint64 `json:"balance"`
}

type Snapshot struct {
	Tokens     []Token `json:"tokens"`
	Events     []Event `json:"events"`
	Allowances []AllowanceRecord `json:"allowances"`
}

type AllowanceRecord struct {
	TokenAddress string `json:"token_address"`
	Owner        string `json:"owner"`
	Spender      string `json:"spender"`
	Amount       uint64 `json:"amount"`
}

type Manager struct {
	mu         sync.RWMutex
	state      *vm.State
	machine    *vm.VM
	tokens     map[string]*Token
	order      []string
	events     []Event
	allowances map[string]map[string]map[string]uint64
}

func NewManager() *Manager {
	state := vm.NewState()
	return &Manager{
		state:      state,
		machine:    vm.New(state),
		tokens:     make(map[string]*Token),
		order:      make([]string, 0),
		events:     make([]Event, 0),
		allowances: make(map[string]map[string]map[string]uint64),
	}
}

func CurrentSpec() Spec {
	return Spec{
		Name:          StandardName,
		Version:       "v0.1-devnet-foundation",
		Runtime:       "PhoenixChain PCVM devnet runtime",
		StandardBadge: StandardName,
		RequiredFields: []string{
			"name", "symbol", "decimals", "totalSupply", "balanceOf",
			"transfer", "approve", "allowance", "transferFrom",
		},
		RequiredEvents: []string{"Transfer", "Approval", "Mint", "Burn", "Paused", "Unpaused"},
		OptionalCaps:   []string{"mint", "burn", "owner/admin", "pause"},
		IntegrityNotes: []string{
			"Current public runtime is devnet, not mainnet.",
			"Current PCVM bytecode template stores canonical supply slot and deploy logs.",
			"Selector-aware public contract call plumbing is not yet exposed through chain transactions.",
			"Registry, holders and transfer history below are real PHX-20 runtime records, not fabricated market data.",
		},
	}
}

func Templates() []Template {
	return []Template{
		{ID: "phx20-basic", Name: "PHX20 Basic", Description: "Fixed-supply fungible token with PHX-20 metadata and transfer registry.", GasEstimate: 78000},
		{ID: "phx20-mintable", Name: "PHX20 Mintable", Description: "Owner-controlled mint extension enabled in devnet runtime.", Mintable: true, GasEstimate: 82000},
		{ID: "phx20-burnable", Name: "PHX20 Burnable", Description: "Supply burn support enabled in devnet runtime.", Burnable: true, GasEstimate: 81000},
		{ID: "phx20-pausable", Name: "PHX20 Pausable", Description: "Emergency transfer pause hooks enabled in devnet runtime.", Pausable: true, GasEstimate: 83000},
		{ID: "phx20-governance", Name: "PHX20 Governance-ready", Description: "Governance-oriented configuration with mint, burn and pause hooks.", Mintable: true, Burnable: true, Pausable: true, Governance: true, GasEstimate: 86000},
	}
}

func (m *Manager) LoadSnapshot(snapshot Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := vm.NewState()
	machine := vm.New(state)
	tokens := make(map[string]*Token, len(snapshot.Tokens))
	order := make([]string, 0, len(snapshot.Tokens))

	for _, token := range snapshot.Tokens {
		template := findTemplate(token.Template)
		nonce := token.DeployNonce
		if nonce == 0 {
			nonce = uint64(token.CreatedAt)
		}
		receipt, err := machine.Deploy(templateCode(token.TotalSupply), token.Creator, nonce, template.GasEstimate)
		if err != nil {
			return fmt.Errorf("restore token %s: %w", token.Symbol, err)
		}
		_ = receipt
		copyToken := token
		if copyToken.Holders == nil {
			copyToken.Holders = make(map[string]uint64)
		}
		tokens[token.Address] = &copyToken
		order = append(order, token.Address)
	}

	allowances := make(map[string]map[string]map[string]uint64)
	for _, item := range snapshot.Allowances {
		if allowances[item.TokenAddress] == nil {
			allowances[item.TokenAddress] = make(map[string]map[string]uint64)
		}
		if allowances[item.TokenAddress][item.Owner] == nil {
			allowances[item.TokenAddress][item.Owner] = make(map[string]uint64)
		}
		allowances[item.TokenAddress][item.Owner][item.Spender] = item.Amount
	}

	m.state = state
	m.machine = machine
	m.tokens = tokens
	m.order = order
	m.events = append([]Event(nil), snapshot.Events...)
	m.allowances = allowances
	return nil
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokens := make([]Token, 0, len(m.order))
	for _, address := range m.order {
		if token, ok := m.tokens[address]; ok {
			copyToken := *token
			copyToken.Holders = cloneHolders(token.Holders)
			tokens = append(tokens, copyToken)
		}
	}
	allowances := make([]AllowanceRecord, 0)
	for tokenAddress, byOwner := range m.allowances {
		for owner, bySpender := range byOwner {
			for spender, amount := range bySpender {
				if amount == 0 {
					continue
				}
				allowances = append(allowances, AllowanceRecord{
					TokenAddress: tokenAddress,
					Owner:        owner,
					Spender:      spender,
					Amount:       amount,
				})
			}
		}
	}
	return Snapshot{
		Tokens:     tokens,
		Events:     append([]Event(nil), m.events...),
		Allowances: allowances,
	}
}

func (m *Manager) Deploy(config DeployConfig) (Token, vm.Receipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := validateDeployConfig(config); err != nil {
		return Token{}, vm.Receipt{}, err
	}
	template := findTemplate(config.Template)
	now := time.Now().UTC()
	deployNonce := uint64(now.UnixNano())
	receipt, err := m.machine.Deploy(templateCode(config.Supply), config.Creator, deployNonce, template.GasEstimate)
	if err != nil {
		return Token{}, receipt, err
	}
	address := strings.ToLower(string(receipt.ContractAddress))
	if _, exists := m.tokens[address]; exists {
		return Token{}, receipt, fmt.Errorf("duplicate token address generated; retry deploy")
	}
	deployTxHash := phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:deploy:%s:%s:%d:%s", address, config.Creator, deployNonce, config.Symbol)))
	token := &Token{
		Address:          address,
		Creator:          strings.ToLower(config.Creator),
		Owner:            strings.ToLower(config.Owner),
		Treasury:         strings.ToLower(config.Treasury),
		Name:             strings.TrimSpace(config.Name),
		Symbol:           strings.ToUpper(strings.TrimSpace(config.Symbol)),
		Decimals:         config.Decimals,
		TotalSupply:      config.Supply,
		DeployTxHash:     deployTxHash,
		DeployNonce:      deployNonce,
		Verified:         false,
		CreatedAt:        now.Unix(),
		StandardBadge:    StandardName,
		RuntimeStatus:    "pcvm devnet token runtime",
		AuditStatus:      "unaudited devnet standard foundation",
		Template:         template.ID,
		LogoURL:          sanitizeURL(config.LogoURL),
		Description:      sanitizeDescription(config.Description),
		WebsiteURL:       sanitizeURL(config.WebsiteURL),
		TwitterURL:       sanitizeURL(config.TwitterURL),
		GithubURL:        sanitizeURL(config.GithubURL),
		Mintable:         config.Mintable || template.Mintable,
		Burnable:         config.Burnable || template.Burnable,
		Pausable:         config.Pausable || template.Pausable,
		Paused:           false,
		DeployGasUsed:    receipt.GasUsed,
		DeployLogCount:   len(receipt.Logs),
		InteractionCount: 1,
		Holders:          map[string]uint64{strings.ToLower(config.Treasury): config.Supply},
	}
	m.tokens[address] = token
	m.order = append([]string{address}, m.order...)
	m.appendEventLocked(Event{
		TokenAddress: address,
		Type:         "Mint",
		TxHash:       deployTxHash,
		From:         ZeroAddress,
		To:           token.Treasury,
		Amount:       config.Supply,
		CreatedAt:    now.Unix(),
		Label:        "Initial PHX-20 supply minted during deploy",
	})
	m.appendEventLocked(Event{
		TokenAddress: address,
		Type:         "Transfer",
		TxHash:       deployTxHash,
		From:         ZeroAddress,
		To:           token.Treasury,
		Amount:       config.Supply,
		CreatedAt:    now.Unix(),
		Label:        "Initial supply allocation",
	})
	return *token, receipt, nil
}

func (m *Manager) Approve(tokenAddress, owner, spender string, amount uint64) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return Event{}, fmt.Errorf("token not found")
	}
	if m.allowances[token.Address] == nil {
		m.allowances[token.Address] = make(map[string]map[string]uint64)
	}
	if m.allowances[token.Address][strings.ToLower(owner)] == nil {
		m.allowances[token.Address][strings.ToLower(owner)] = make(map[string]uint64)
	}
	m.allowances[token.Address][strings.ToLower(owner)][strings.ToLower(spender)] = amount
	token.InteractionCount++
	event := Event{
		TokenAddress: token.Address,
		Type:         "Approval",
		TxHash:       phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:approve:%s:%s:%s:%d:%d", token.Address, owner, spender, amount, time.Now().UTC().UnixNano()))),
		Owner:        strings.ToLower(owner),
		Spender:      strings.ToLower(spender),
		Amount:       amount,
		CreatedAt:    time.Now().UTC().Unix(),
		Label:        "Allowance updated",
	}
	m.appendEventLocked(event)
	return event, nil
}

func (m *Manager) Allowance(tokenAddress, owner, spender string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.allowances[strings.ToLower(tokenAddress)][strings.ToLower(owner)][strings.ToLower(spender)]
}

func (m *Manager) Transfer(tokenAddress, from, to string, amount uint64) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.transferLocked(tokenAddress, from, to, amount, "Transfer")
}

func (m *Manager) TransferFrom(tokenAddress, owner, spender, to string, amount uint64) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tokenAddress = strings.ToLower(tokenAddress)
	owner = strings.ToLower(owner)
	spender = strings.ToLower(spender)
	to = strings.ToLower(to)
	if m.allowances[tokenAddress] == nil || m.allowances[tokenAddress][owner] == nil || m.allowances[tokenAddress][owner][spender] < amount {
		return Event{}, fmt.Errorf("allowance exceeded")
	}
	event, err := m.transferLocked(tokenAddress, owner, to, amount, "Transfer")
	if err != nil {
		return Event{}, err
	}
	m.allowances[tokenAddress][owner][spender] -= amount
	return event, nil
}

func (m *Manager) Mint(tokenAddress, caller, to string, amount uint64) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return Event{}, fmt.Errorf("token not found")
	}
	if !token.Mintable {
		return Event{}, fmt.Errorf("mint disabled for this token")
	}
	if strings.ToLower(caller) != token.Owner {
		return Event{}, fmt.Errorf("only owner can mint")
	}
	token.Holders[strings.ToLower(to)] += amount
	token.TotalSupply += amount
	token.InteractionCount++
	event := Event{
		TokenAddress: token.Address,
		Type:         "Mint",
		TxHash:       phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:mint:%s:%s:%s:%d:%d", token.Address, caller, to, amount, time.Now().UTC().UnixNano()))),
		From:         ZeroAddress,
		To:           strings.ToLower(to),
		Amount:       amount,
		CreatedAt:    time.Now().UTC().Unix(),
		Label:        "Owner mint executed",
	}
	m.appendEventLocked(event)
	m.appendEventLocked(Event{
		TokenAddress: token.Address,
		Type:         "Transfer",
		TxHash:       event.TxHash,
		From:         ZeroAddress,
		To:           strings.ToLower(to),
		Amount:       amount,
		CreatedAt:    event.CreatedAt,
		Label:        "Mint transfer",
	})
	return event, nil
}

func (m *Manager) Burn(tokenAddress, caller string, amount uint64) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return Event{}, fmt.Errorf("token not found")
	}
	if !token.Burnable {
		return Event{}, fmt.Errorf("burn disabled for this token")
	}
	caller = strings.ToLower(caller)
	if token.Holders[caller] < amount {
		return Event{}, fmt.Errorf("insufficient token balance")
	}
	token.Holders[caller] -= amount
	if token.Holders[caller] == 0 {
		delete(token.Holders, caller)
	}
	token.TotalSupply -= amount
	token.InteractionCount++
	event := Event{
		TokenAddress: token.Address,
		Type:         "Burn",
		TxHash:       phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:burn:%s:%s:%d:%d", token.Address, caller, amount, time.Now().UTC().UnixNano()))),
		From:         caller,
		To:           ZeroAddress,
		Amount:       amount,
		CreatedAt:    time.Now().UTC().Unix(),
		Label:        "Token burn executed",
	}
	m.appendEventLocked(event)
	m.appendEventLocked(Event{
		TokenAddress: token.Address,
		Type:         "Transfer",
		TxHash:       event.TxHash,
		From:         caller,
		To:           ZeroAddress,
		Amount:       amount,
		CreatedAt:    event.CreatedAt,
		Label:        "Burn transfer",
	})
	return event, nil
}

func (m *Manager) SetPaused(tokenAddress, caller string, paused bool) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return Event{}, fmt.Errorf("token not found")
	}
	if !token.Pausable {
		return Event{}, fmt.Errorf("pause disabled for this token")
	}
	if strings.ToLower(caller) != token.Owner {
		return Event{}, fmt.Errorf("only owner can change pause state")
	}
	token.Paused = paused
	token.InteractionCount++
	eventType := "Unpaused"
	label := "Token transfers resumed"
	if paused {
		eventType = "Paused"
		label = "Token transfers paused by owner"
	}
	event := Event{
		TokenAddress: token.Address,
		Type:         eventType,
		TxHash:       phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:pause:%s:%s:%t:%d", token.Address, caller, paused, time.Now().UTC().UnixNano()))),
		Owner:        token.Owner,
		CreatedAt:    time.Now().UTC().Unix(),
		Label:        label,
	}
	m.appendEventLocked(event)
	return event, nil
}

func (m *Manager) ListTokens() []TokenSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]TokenSummary, 0, len(m.order))
	for _, address := range m.order {
		token := m.tokens[address]
		out = append(out, summarizeToken(*token, countTransfers(m.events, address)))
	}
	return out
}

func (m *Manager) QueryTokens(query, sortBy string) []TokenSummary {
	out := m.ListTokens()
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]TokenSummary, 0, len(out))
	for _, item := range out {
		if query == "" ||
			strings.Contains(strings.ToLower(item.Name), query) ||
			strings.Contains(strings.ToLower(item.Symbol), query) ||
			strings.Contains(strings.ToLower(item.Address), query) ||
			strings.Contains(strings.ToLower(item.Creator), query) {
			filtered = append(filtered, item)
		}
	}
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "name":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Name < filtered[j].Name })
	case "supply":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].TotalSupply > filtered[j].TotalSupply })
	case "holders":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].HoldersCount > filtered[j].HoldersCount })
	case "transfers":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].TransferCount > filtered[j].TransferCount })
	default:
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt > filtered[j].CreatedAt })
	}
	return filtered
}

func (m *Manager) Token(address string) (Token, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.tokens[strings.ToLower(address)]
	if !ok {
		return Token{}, false
	}
	copyToken := *token
	copyToken.Holders = cloneHolders(token.Holders)
	return copyToken, true
}

func (m *Manager) Holders(address string) []HolderBalance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.tokens[strings.ToLower(address)]
	if !ok {
		return nil
	}
	holders := make([]HolderBalance, 0, len(token.Holders))
	for holder, balance := range token.Holders {
		holders = append(holders, HolderBalance{Address: holder, Balance: balance})
	}
	sort.Slice(holders, func(i, j int) bool {
		if holders[i].Balance == holders[j].Balance {
			return holders[i].Address < holders[j].Address
		}
		return holders[i].Balance > holders[j].Balance
	})
	return holders
}

func (m *Manager) EventsByToken(address string) []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Event, 0)
	for _, event := range m.events {
		if event.TokenAddress == strings.ToLower(address) {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

func (m *Manager) AddressBalances(address string) []WalletBalance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	address = strings.ToLower(address)
	out := make([]WalletBalance, 0)
	for _, tokenAddress := range m.order {
		token := m.tokens[tokenAddress]
		balance := token.Holders[address]
		if balance == 0 {
			continue
		}
		out = append(out, WalletBalance{
			TokenAddress: token.Address,
			Name:         token.Name,
			Symbol:       token.Symbol,
			Decimals:     token.Decimals,
			Balance:      balance,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Balance > out[j].Balance })
	return out
}

func (m *Manager) BalanceOf(tokenAddress, address string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return 0
	}
	return token.Holders[strings.ToLower(address)]
}

func (m *Manager) TokenState() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return map[string]any{
		"supported":              true,
		"status":                 "phx20-devnet-foundation-live",
		"registry_count":         len(m.tokens),
		"indexed_transfer_count": countTransferEvents(m.events),
		"note":                   "PHX-20 is live as a PCVM-backed devnet token foundation. Public chain transaction calldata plumbing is still maturing.",
	}
}

func (m *Manager) transferLocked(tokenAddress, from, to string, amount uint64, eventType string) (Event, error) {
	token, ok := m.tokens[strings.ToLower(tokenAddress)]
	if !ok {
		return Event{}, fmt.Errorf("token not found")
	}
	if token.Paused {
		return Event{}, fmt.Errorf("token is paused")
	}
	from = strings.ToLower(from)
	to = strings.ToLower(to)
	if !isAddress(from) || !isAddress(to) {
		return Event{}, fmt.Errorf("invalid from/to address")
	}
	if amount == 0 {
		return Event{}, fmt.Errorf("amount must be greater than zero")
	}
	if token.Holders[from] < amount {
		return Event{}, fmt.Errorf("insufficient token balance")
	}
	token.Holders[from] -= amount
	if token.Holders[from] == 0 {
		delete(token.Holders, from)
	}
	token.Holders[to] += amount
	token.InteractionCount++
	event := Event{
		TokenAddress: token.Address,
		Type:         eventType,
		TxHash:       phxcrypto.HashBytes([]byte(fmt.Sprintf("phx20:transfer:%s:%s:%s:%d:%d", token.Address, from, to, amount, time.Now().UTC().UnixNano()))),
		From:         from,
		To:           to,
		Amount:       amount,
		CreatedAt:    time.Now().UTC().Unix(),
		Label:        "PHX-20 transfer recorded",
	}
	m.appendEventLocked(event)
	return event, nil
}

func (m *Manager) appendEventLocked(event Event) {
	event.ID = phxcrypto.HashBytes([]byte(fmt.Sprintf("%s:%s:%d", event.TokenAddress, event.TxHash, len(m.events))))
	m.events = append(m.events, event)
}

func validateDeployConfig(config DeployConfig) error {
	config.Name = strings.TrimSpace(config.Name)
	config.Symbol = strings.TrimSpace(config.Symbol)
	if !isAddress(config.Creator) {
		return fmt.Errorf("creator must be a valid Phoenix address")
	}
	if !isAddress(config.Owner) {
		return fmt.Errorf("owner must be a valid Phoenix address")
	}
	if !isAddress(config.Treasury) {
		return fmt.Errorf("treasury must be a valid Phoenix address")
	}
	if config.Name == "" || len(config.Name) > MaxNameLength {
		return fmt.Errorf("name must be 1-%d chars", MaxNameLength)
	}
	if config.Symbol == "" || len(config.Symbol) > MaxSymbolLength {
		return fmt.Errorf("symbol must be 1-%d chars", MaxSymbolLength)
	}
	for _, ch := range config.Symbol {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			return fmt.Errorf("symbol must be alphanumeric")
		}
	}
	if config.Decimals > MaxDecimals {
		return fmt.Errorf("decimals must be <= %d", MaxDecimals)
	}
	if config.Supply == 0 {
		return fmt.Errorf("supply must be greater than zero")
	}
	if len(strings.TrimSpace(config.Description)) > MaxDescriptionLength {
		return fmt.Errorf("description must be <= %d chars", MaxDescriptionLength)
	}
	for _, value := range []string{config.LogoURL, config.WebsiteURL, config.TwitterURL, config.GithubURL} {
		if value == "" {
			continue
		}
		if !isSafeURL(value) {
			return fmt.Errorf("metadata urls must use http or https")
		}
		if len(strings.TrimSpace(value)) > MaxURLLength {
			return fmt.Errorf("metadata urls must be <= %d chars", MaxURLLength)
		}
	}
	if findTemplate(config.Template).ID == "" {
		return fmt.Errorf("unknown template")
	}
	return nil
}

func isAddress(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 43 || !strings.HasPrefix(value, "phx") {
		return false
	}
	for _, ch := range value[3:] {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

func findTemplate(id string) Template {
	for _, item := range Templates() {
		if item.ID == strings.TrimSpace(strings.ToLower(id)) {
			return item
		}
	}
	return Template{}
}

func templateCode(supply uint64) []byte {
	return vm.BasicTokenExample(supply)
}

func summarizeToken(token Token, transferCount int) TokenSummary {
	return TokenSummary{
		Address:       token.Address,
		Creator:       token.Creator,
		Name:          token.Name,
		Symbol:        token.Symbol,
		Decimals:      token.Decimals,
		TotalSupply:   token.TotalSupply,
		DeployTxHash:  token.DeployTxHash,
		Verified:      token.Verified,
		CreatedAt:     token.CreatedAt,
		StandardBadge: token.StandardBadge,
		RuntimeStatus: token.RuntimeStatus,
		AuditStatus:   token.AuditStatus,
		Template:      token.Template,
		LogoURL:       token.LogoURL,
		Description:   token.Description,
		WebsiteURL:    token.WebsiteURL,
		TwitterURL:    token.TwitterURL,
		GithubURL:     token.GithubURL,
		Mintable:      token.Mintable,
		Burnable:      token.Burnable,
		Pausable:      token.Pausable,
		Paused:        token.Paused,
		HoldersCount:  len(token.Holders),
		TransferCount: transferCount,
	}
}

func countTransferEvents(events []Event) int {
	total := 0
	for _, event := range events {
		if event.Type == "Transfer" {
			total++
		}
	}
	return total
}

func countTransfers(events []Event, tokenAddress string) int {
	total := 0
	for _, event := range events {
		if event.TokenAddress == tokenAddress && event.Type == "Transfer" {
			total++
		}
	}
	return total
}

func cloneHolders(in map[string]uint64) map[string]uint64 {
	out := make(map[string]uint64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func SpecJSON() string {
	data, _ := json.MarshalIndent(CurrentSpec(), "", "  ")
	return string(data)
}

func isSafeURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
}

func sanitizeURL(value string) string {
	value = strings.TrimSpace(value)
	if !isSafeURL(value) {
		return ""
	}
	if len(value) > MaxURLLength {
		return value[:MaxURLLength]
	}
	return value
}

func sanitizeDescription(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > MaxDescriptionLength {
		return value[:MaxDescriptionLength]
	}
	return value
}

func (t Token) CreatedLabel() string {
	if t.CreatedAt <= 0 {
		return "runtime-unknown"
	}
	return time.Unix(t.CreatedAt, 0).UTC().Format(time.RFC3339)
}

func (t TokenSummary) CreatedLabel() string {
	if t.CreatedAt <= 0 {
		return "runtime-unknown"
	}
	return time.Unix(t.CreatedAt, 0).UTC().Format(time.RFC3339)
}

func (t TokenSummary) CapabilitySummary() string {
	values := make([]string, 0, 4)
	if t.Mintable {
		values = append(values, "mint")
	}
	if t.Burnable {
		values = append(values, "burn")
	}
	if t.Pausable {
		values = append(values, "pause")
	}
	if strings.Contains(t.Template, "governance") {
		values = append(values, "governance")
	}
	if len(values) == 0 {
		return "basic"
	}
	return strings.Join(values, ", ")
}

func (t Token) CapabilitySummary() string {
	return summarizeToken(t, 0).CapabilitySummary()
}

func ParseBoolFlag(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

func ParseSortPage(value string) int {
	out, _ := strconv.Atoi(strings.TrimSpace(value))
	if out <= 0 {
		return 1
	}
	return out
}
