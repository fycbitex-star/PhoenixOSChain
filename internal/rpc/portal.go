package rpc

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/indexer"
)

func (s *Server) handleExplorerAlias(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Explorer")
	renderExplorer(w, `{{define "content"}}
<section class="hero"><div class="panel"><h1>PhoenixOS Chain Explorer</h1><p class="muted">Explore blocks, transactions, validators, faucet activity and live runtime state from a single chain portal surface.</p><form class="portal-search" method="get" action="/search"><input type="text" name="q" placeholder="Search block, tx, address or contract"><button type="submit">Search Explorer</button></form></div><div class="panel"><h3>Live Network Summary</h3><p class="hash">RPC: https://chain.phoenixos.us</p><p>Chain ID: {{.ChainID}}</p><p>Consensus: PoA Devnet</p><p>Native asset: {{index .NativeAsset "max_supply_compact"}}</p></div></section>
<section class="grid"><div class="metric"><div class="label">Latest Block</div><div class="value">{{.Head.Height}}</div></div><div class="metric"><div class="label">Recent Transactions</div><div class="value">{{len .Transactions}}</div></div><div class="metric"><div class="label">Validators</div><div class="value">{{.ValidatorCount}}</div></div><div class="metric"><div class="label">Healthy Peers</div><div class="value">{{.HealthyPeerCount}}</div></div><div class="metric"><div class="label">Max Supply</div><div class="value">{{index .NativeAsset "max_supply_compact"}}</div></div><div class="metric"><div class="label">Circulating</div><div class="value">{{index .NativeAsset "circulating_note"}}</div></div></section>
<div class="section-title"><h2>Latest Blocks</h2><a href="/blocks">Full list</a></div>{{template "blocksTable" .}}
<div class="section-title"><h2>Latest Transactions</h2><a href="/tx">Full list</a></div>{{template "txTable" .}}
{{template "networkReadiness" .}}
{{end}}`, data)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	eventType, normalized := indexer.NormalizeSearchQuery(query)
	switch eventType {
	case "address":
		http.Redirect(w, r, "/address/"+url.PathEscape(normalized), http.StatusSeeOther)
		return
	case "contract":
		if _, ok := s.phx20.Token(normalized); ok {
			http.Redirect(w, r, "/tokens/"+url.PathEscape(normalized), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/contracts/"+url.PathEscape(normalized), http.StatusSeeOther)
		return
	case "block_height":
		height, _ := strconv.ParseUint(normalized, 10, 64)
		if _, ok := s.indexer.FindBlockByHeight(height); ok {
			http.Redirect(w, r, fmt.Sprintf("/blocks/%d", height), http.StatusSeeOther)
			return
		}
	case "hash":
		if txValue, ok := s.indexer.FindTransaction(normalized); ok {
			if prefersHTML(r) {
				http.Redirect(w, r, "/tx/"+url.PathEscape(txValue.Hash), http.StatusSeeOther)
				return
			}
			writeJSON(w, http.StatusOK, txValue)
			return
		}
		if block, ok := s.indexer.FindBlockByHash(normalized); ok {
			http.Redirect(w, r, fmt.Sprintf("/blocks/%d", block.Height), http.StatusSeeOther)
			return
		}
	}

	data := s.explorerData("PhoenixOS Chain Search")
	data["SearchQuery"] = query
	data["SearchMiss"] = true
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Search</h1><span class="muted">Universal chain lookup</span></div>
<div class="panel">
<p class="muted">We could not resolve <span class="hash">{{.SearchQuery}}</span> as a known block height, block hash, transaction hash or Phoenix address in the current devnet runtime.</p>
<p>This search currently supports:</p>
<ul class="portal-list">
<li>block number</li>
<li>block hash</li>
<li>transaction hash</li>
<li>Phoenix address</li>
<li>contract address</li>
</ul>
</div>
{{template "networkReadiness" .}}
{{end}}`, data)
}

func (s *Server) handleTxListAlias(w http.ResponseWriter, r *http.Request) {
	s.handleExplorerTxs(w, r)
}

func (s *Server) handleAddressPage(w http.ResponseWriter, r *http.Request) {
	address, ok := normalizeAddress(r.PathValue("address"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid address")
		return
	}

	account, activityPage, ok := s.indexer.AddressActivity(address, pageFromRequest(r), pageSizeFromRequest(r))
	if !ok {
		writeError(w, http.StatusNotFound, "address not indexed")
		return
	}
	data := s.explorerData("PhoenixOS Chain Address")
	data["Address"] = address
	data["Account"] = account
	data["AddressTransactions"] = activityPage.Items
	data["TxPage"] = activityPage
	data["Transactions"] = data["AddressTransactions"]
	data["TokenBalances"] = s.phx20.AddressBalances(address)
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Address</h1><span class="muted">Account, transfers and runtime footprint</span></div>
<div class="grid">
  <div class="metric"><div class="label">Address</div><div class="value hash">{{short .Address}}</div></div>
  <div class="metric"><div class="label">Balance</div><div class="value">{{.Account.Balance}} PHX</div></div>
  <div class="metric"><div class="label">Nonce</div><div class="value">{{.Account.Nonce}}</div></div>
  <div class="metric"><div class="label">Indexed Activity</div><div class="value">{{.Account.ActivityCounter.Transactions}}</div></div>
</div>
<div class="section-title"><h2>Overview</h2><span class="muted">Current runtime account state</span></div>
<div class="panel">
  <p class="hash">Address: {{.Address}}</p>
  <p>Transfers sent: {{.Account.ActivityCounter.TransfersSent}} | Transfers received: {{.Account.ActivityCounter.TransfersReceived}}</p>
  <p>Contract calls: {{.Account.ActivityCounter.ContractCalls}} | Token transfers: {{.Account.ActivityCounter.TokenTransfers}}</p>
  <p>Wallet connect actions are planned inside PhoenixOS and are not yet exposed directly on this explorer.</p>
  <p class="muted">Native balance and tx history below are indexer-backed and real. PHX-20 balances are loaded from the live token registry.</p>
</div>
<div class="section-title"><h2>PHX-20 Balances</h2><span class="muted">Token registry holdings for this address</span></div>
<table><tr><th>Token</th><th>Symbol</th><th>Balance</th><th>Decimals</th></tr>{{range .TokenBalances}}<tr><td><a href="/tokens/{{.TokenAddress}}">{{.Name}}</a></td><td>{{.Symbol}}</td><td>{{.Balance}}</td><td>{{.Decimals}}</td></tr>{{else}}<tr><td colspan="4" class="muted">No PHX-20 balances indexed for this address yet.</td></tr>{{end}}</table>
<div class="section-title"><h2>Transactions</h2><span class="muted">Most recent related transactions</span></div>
{{template "txTable" .}}
{{template "pageState" .TxPage}}
{{template "modulePreview" dict "Title" "Address analytics roadmap" "Body" "Labels, QR export, event tabs and richer contract interaction analytics still require a deeper historical indexer. The balances and transactions shown above are real runtime data."}}
{{end}}`, data)
}

func (s *Server) handleBlockRoute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var (
		block indexer.BlockRecord
		ok    bool
	)
	if height, err := strconv.ParseUint(id, 10, 64); err == nil {
		block, ok = s.indexer.FindBlockByHeight(height)
	} else {
		block, ok = s.indexer.FindBlockByHash(strings.ToLower(id))
	}
	if !ok {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}
	s.renderBlockDetail(w, block)
}

func (s *Server) handleTxRoute(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimSpace(strings.ToLower(r.PathValue("hash")))
	transaction, ok := s.indexer.FindTransaction(hash)
	if !ok {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	if !prefersHTML(r) {
		writeJSON(w, http.StatusOK, transaction)
		return
	}
	s.renderTransactionDetail(w, transaction)
}

func (s *Server) handleContractsIndex(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Contracts")
	contracts := s.indexer.ContractsPage(pageFromRequest(r), pageSizeFromRequest(r))
	data["ContractPage"] = contracts
	data["Contracts"] = contracts.Items
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Contracts</h1><span class="muted">Runtime and verification layer</span></div>
<div class="panel">
  <p>PhoenixChain includes an internal VM prototype and PCVM-facing building blocks in source. Below is the live interaction index observed by the explorer runtime.</p>
  <p class="muted">Verified source, ABI decode and public contract write/read flows are not faked before the supporting runtime exists.</p>
</div>
<table><tr><th>Address</th><th>Type</th><th>Runtime</th><th>Interactions</th></tr>{{range .Contracts}}<tr><td class="hash"><a href="/contracts/{{.Address}}">{{short .Address}}</a></td><td>{{.ContractType}}</td><td>{{.RuntimeStatus}}</td><td>{{.InteractionCount}}</td></tr>{{else}}<tr><td colspan="4" class="muted">No indexed contracts yet.</td></tr>{{end}}</table>
{{template "pageState" .ContractPage}}
{{template "modulePreview" dict "Title" "Contract runtime readiness" "Body" "Available in code: VM modules, contract data structures and token example prototypes. Missing for public portal: verified source registry, contract metadata index, event decoding, ABI surfaces and write/read contract UX."}}
{{end}}`, data)
}

func (s *Server) handleContractDetail(w http.ResponseWriter, r *http.Request) {
	address := strings.TrimSpace(strings.ToLower(r.PathValue("address")))
	data := s.explorerData("PhoenixOS Chain Contract")
	data["ContractAddress"] = address
	if contract, ok := s.indexer.FindContract(address); ok {
		data["Contract"] = contract
		data["Transactions"] = contract.RecentInteractions
		renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Contract</h1><span class="muted">Indexed runtime intelligence</span></div>
<section class="grid">
  <div class="metric"><div class="label">Address</div><div class="value hash">{{short .Contract.Address}}</div></div>
  <div class="metric"><div class="label">Runtime</div><div class="value">{{.Contract.RuntimeStatus}}</div></div>
  <div class="metric"><div class="label">Type</div><div class="value">{{.Contract.ContractType}}</div></div>
  <div class="metric"><div class="label">Interactions</div><div class="value">{{.Contract.InteractionCount}}</div></div>
</section>
<div class="panel">
  <p class="hash">Contract address: {{.Contract.Address}}</p>
  <p>Creator: {{if .Contract.Creator}}{{.Contract.Creator}}{{else}}<span class="warn">not indexed yet</span>{{end}}</p>
  <p>Creation tx: {{if .Contract.CreationTxHash}}<a href="/tx/{{.Contract.CreationTxHash}}">{{short .Contract.CreationTxHash}}</a>{{else}}<span class="warn">not indexed yet</span>{{end}}</p>
  <p>ABI: <span class="warn">{{if .Contract.ABIAvailable}}available{{else}}runtime unavailable{{end}}</span></p>
  <p>Verified source: <span class="warn">{{if .Contract.VerifiedSource}}verified{{else}}not verified in current runtime{{end}}</span></p>
  <p>Events/logs: <span class="warn">feature pending runtime support</span></p>
</div>
<div class="section-title"><h2>Recent Calls</h2><span class="muted">Observed contract interactions</span></div>
{{template "txTable" .}}
{{template "modulePreview" dict "Title" "Contract intelligence roadmap" "Body" "Next layer: ABI store, function selectors, calldata decode, receipt/log indexing, source verification and read/write contract panels."}}
{{end}}`, data)
		return
	}
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Contract</h1><span class="muted">Preview state with technical integrity</span></div>
<div class="panel">
  <p class="hash">Requested contract: {{.ContractAddress}}</p>
  <p>This contract has not been observed by the current runtime indexer yet.</p>
  <p class="muted">Current runtime status: VM prototype exists in source; observed interactions are indexed when they reach the chain.</p>
</div>
{{template "modulePreview" dict "Title" "What will unlock this page" "Body" "Contract deployment registry, ABI store, bytecode hash indexing, event/log decoding, source verification flow, contract call explorer, token standard maturity."}}
{{end}}`, data)
}

func (s *Server) handleTokensIndex(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Tokens")
	data["TokenState"] = s.phx20.TokenState()
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortBy == "" {
		sortBy = "created"
	}
	data["Tokens"] = s.phx20.QueryTokens(query, sortBy)
	data["TokenQuery"] = query
	data["TokenSort"] = sortBy
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Tokens</h1><span class="muted">Network asset registry</span></div>
<div class="panel">
  <p>PHX-20 is the official fungible token foundation for PhoenixChain devnet. Tokens listed here are real PCVM-backed registry entries created through the PHX-20 runtime manager.</p>
  <p class="muted">{{.TokenState.Note}}</p>
  <p>Indexed transfer count: {{index .TokenState "indexed_transfer_count"}} | Runtime status: {{index .TokenState "status"}} | Registry count: {{index .TokenState "registry_count"}}</p>
</div>
<div class="portal-actions" style="margin-bottom:12px"><a class="portal-button" href="/tokens/factory">Open Token Factory</a><a class="portal-button secondary" href="/api/phx20/spec">PHX-20 Spec JSON</a></div>
<form class="portal-actions" method="get" action="/tokens" style="margin-bottom:12px">
  <input class="portal-input" name="q" value="{{.TokenQuery}}" placeholder="Search name, symbol, address or creator">
  <select class="portal-input" name="sort" style="max-width:220px">
    <option value="created" {{if eq .TokenSort "created"}}selected{{end}}>Newest</option>
    <option value="name" {{if eq .TokenSort "name"}}selected{{end}}>Name</option>
    <option value="supply" {{if eq .TokenSort "supply"}}selected{{end}}>Supply</option>
    <option value="holders" {{if eq .TokenSort "holders"}}selected{{end}}>Holders</option>
    <option value="transfers" {{if eq .TokenSort "transfers"}}selected{{end}}>Transfers</option>
  </select>
  <button class="portal-button" type="submit">Apply</button>
</form>
<table><tr><th>Name</th><th>Symbol</th><th>Address</th><th>Creator</th><th>Supply</th><th>Holders</th><th>Transfers</th><th>Verified</th><th>Created</th><th>Standard</th></tr>{{range .Tokens}}<tr><td><a href="/tokens/{{.Address}}">{{.Name}}</a></td><td>{{.Symbol}}</td><td class="hash"><a href="/tokens/{{.Address}}">{{short .Address}}</a></td><td class="hash"><a href="/address/{{.Creator}}">{{short .Creator}}</a></td><td>{{.TotalSupply}}</td><td>{{.HoldersCount}}</td><td>{{.TransferCount}}</td><td>{{.Verified}}</td><td>{{.CreatedLabel}}</td><td><span class="tag">{{.StandardBadge}}</span></td></tr>{{else}}<tr><td colspan="10" class="muted">No PHX-20 tokens matched this query.</td></tr>{{end}}</table>
{{template "modulePreview" dict "Title" "Token portal readiness" "Body" "Live now: PHX-20 registry, token metadata, holders, transfer history and deploy references. Still pending for deeper parity: token price feeds, verification badges, wallet-signed transfer UX and richer token analytics."}}
{{end}}`, data)
}

func (s *Server) handleTokenDetail(w http.ResponseWriter, r *http.Request) {
	address := strings.ToLower(r.PathValue("address"))
	data := s.explorerData("PhoenixOS Chain Token")
	data["TokenAddress"] = address
	if token, ok := s.phx20.Token(address); ok {
		data["Token"] = token
		data["TokenHolders"] = s.phx20.Holders(address)
		data["TokenEvents"] = s.phx20.EventsByToken(address)
		data["DeployStatus"] = strings.TrimSpace(r.URL.Query().Get("status"))
		data["DeployMessage"] = strings.TrimSpace(r.URL.Query().Get("message"))
		renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>{{.Token.Name}}</h1><span class="muted">{{.Token.Symbol}} · {{.Token.StandardBadge}}</span></div>
{{if .DeployStatus}}<div class="panel"><p><strong>{{.DeployStatus}}</strong></p><p>{{.DeployMessage}}</p></div>{{end}}
<section class="grid">
  <div class="metric"><div class="label">Token Address</div><div class="value hash">{{short .Token.Address}}</div></div>
  <div class="metric"><div class="label">Supply</div><div class="value">{{.Token.TotalSupply}}</div></div>
  <div class="metric"><div class="label">Decimals</div><div class="value">{{.Token.Decimals}}</div></div>
  <div class="metric"><div class="label">Runtime</div><div class="value">{{.Token.RuntimeStatus}}</div></div>
  <div class="metric"><div class="label">Template</div><div class="value">{{.Token.Template}}</div></div>
  <div class="metric"><div class="label">Audit</div><div class="value">{{.Token.AuditStatus}}</div></div>
  <div class="metric"><div class="label">Mint / Burn / Pause</div><div class="value">{{.Token.Mintable}} / {{.Token.Burnable}} / {{.Token.Pausable}}</div></div>
  <div class="metric"><div class="label">Paused</div><div class="value">{{.Token.Paused}}</div></div>
</section>
<div class="panel">
  <p class="hash">Creator: <a href="/address/{{.Token.Creator}}">{{.Token.Creator}}</a></p>
  <p class="hash">Owner: <a href="/address/{{.Token.Owner}}">{{.Token.Owner}}</a></p>
  <p class="hash">Treasury: <a href="/address/{{.Token.Treasury}}">{{.Token.Treasury}}</a></p>
  <p class="hash">Deploy reference: {{.Token.DeployTxHash}}</p>
  <p>Verified: {{.Token.Verified}}</p>
  <p>Created at: {{.Token.CreatedAt}}</p>
  <p>Deploy gas used: {{.Token.DeployGasUsed}} | Deploy logs: {{.Token.DeployLogCount}}</p>
  <p class="muted">This deploy reference is a real PHX-20 runtime record in current devnet foundation. Public chain calldata-backed deploy transactions are still maturing.</p>
</div>
<div class="section-title"><h2>Token Metadata</h2><span class="muted">Registry-backed metadata surface</span></div>
<div class="panel">
  <p>Description: {{if .Token.Description}}{{.Token.Description}}{{else}}<span class="warn">not provided</span>{{end}}</p>
  <p>Website: {{if .Token.WebsiteURL}}<a href="{{.Token.WebsiteURL}}">{{.Token.WebsiteURL}}</a>{{else}}<span class="warn">not provided</span>{{end}}</p>
  <p>Social: {{if .Token.TwitterURL}}<a href="{{.Token.TwitterURL}}">{{.Token.TwitterURL}}</a>{{else}}<span class="warn">not provided</span>{{end}}</p>
  <p>GitHub: {{if .Token.GithubURL}}<a href="{{.Token.GithubURL}}">{{.Token.GithubURL}}</a>{{else}}<span class="warn">not provided</span>{{end}}</p>
  <p>Logo: {{if .Token.LogoURL}}<a href="{{.Token.LogoURL}}">{{.Token.LogoURL}}</a>{{else}}<span class="warn">not provided</span>{{end}}</p>
</div>
<div class="section-title"><h2>Holders</h2><span class="muted">Indexed registry balances</span></div>
<table><tr><th>Address</th><th>Balance</th></tr>{{range .TokenHolders}}<tr><td class="hash"><a href="/address/{{.Address}}">{{short .Address}}</a></td><td>{{.Balance}}</td></tr>{{else}}<tr><td colspan="2" class="muted">No holders indexed.</td></tr>{{end}}</table>
<div class="section-title"><h2>Transfers and Events</h2><span class="muted">PHX-20 runtime event stream</span></div>
<table><tr><th>Type</th><th>From</th><th>To</th><th>Amount</th><th>Reference</th></tr>{{range .TokenEvents}}<tr><td>{{.Type}}</td><td class="hash">{{short .From}}</td><td class="hash">{{short .To}}</td><td>{{.Amount}}</td><td class="hash">{{short .TxHash}}</td></tr>{{else}}<tr><td colspan="5" class="muted">No events recorded yet.</td></tr>{{end}}</table>
{{template "modulePreview" dict "Title" "Contract / verification / wallet state" "Body" "Contract tab parity, verification workflows and wallet-signed send/import flows still need deeper runtime adapters. This page shows only live registry metadata and event history already observed by the devnet foundation."}}
{{end}}`, data)
		return
	}
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Token Detail</h1><span class="muted">Registry lookup failed</span></div>
<div class="panel">
  <p class="hash">Requested token: {{.TokenAddress}}</p>
  <p>This PHX-20 token is not in the live registry yet.</p>
</div>
{{template "modulePreview" dict "Title" "What unlocks this page" "Body" "Deploy a token from the PHX-20 factory or let the runtime restore it from snapshot. This page never invents token metadata that the registry does not actually contain."}}
{{end}}`, data)
}

func (s *Server) handleValidatorsPortal(w http.ResponseWriter, r *http.Request) {
	s.handleExplorerValidators(w, r)
}

func (s *Server) handleTokenomicsPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Tokenomics")
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Tokenomics</h1><span class="muted">Native PHX monetary architecture</span></div>
<section class="grid">
  <div class="metric"><div class="label">Native Asset</div><div class="value">{{index .NativeAsset "symbol"}}</div></div>
  <div class="metric"><div class="label">Decimals</div><div class="value">{{index .NativeAsset "decimals"}}</div></div>
  <div class="metric"><div class="label">Max Supply</div><div class="value">{{index .NativeAsset "max_supply_display"}}</div></div>
  <div class="metric"><div class="label">Emission State</div><div class="value">{{index .NativeAsset "emission_state"}}</div></div>
</section>
<div class="panel">
  <p><strong>{{index .NativeAsset "max_supply_display"}}</strong> is the canonical PHX maximum supply reference across explorer, docs, wallet metadata, RPC metadata, governance positioning and future validator economics.</p>
  <p>Display forms: <span class="tag">{{index .NativeAsset "max_supply_compact"}}</span> <span class="tag">{{index .NativeAsset "max_supply_display"}}</span></p>
  <p>Circulating supply: <span class="warn">{{index .NativeAsset "circulating_note"}}</span></p>
  <p class="muted">No fake FDV, fake market cap, fake exchange listing or fake unlocked supply is shown here.</p>
</div>
<div class="section-title"><h2>Allocation Architecture</h2><span class="muted">Placeholder percentages, not live unlock reporting</span></div>
<table><tr><th>Bucket</th><th>Placeholder Share</th><th>Current Note</th></tr>{{range .SupplyAllocations}}<tr><td>{{.Label}}</td><td>{{.Percentage}}</td><td>{{.Scope}}</td></tr>{{end}}</table>
<div class="section-title"><h2>Economic Positioning</h2><span class="muted">Infrastructure-grade asset posture</span></div>
<div class="panel">
  <p>PHX is positioned as a limited-supply infrastructure asset, not a hyper-inflation meme coin.</p>
  <p>Future utility scope: {{range $index, $item := index .NativeAsset "governance_utilities"}}{{if $index}}, {{end}}{{$item}}{{end}}.</p>
  <p class="muted">Validator rewards, staking emissions, treasury schedules and governance unlock logic require dedicated live modules before this page can show operational distribution data.</p>
</div>
{{end}}`, data)
}

func (s *Server) handleStakingPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Staking")
	records := s.governanceAwareValidatorRecords()
	data["ValidatorRecords"] = records
	data["ValidatorSecurity"] = validatorSecurity(records)
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Staking</h1><span class="muted">Validator delegation and reward layer</span></div>
<section class="grid">
  <div class="metric"><div class="label">Total Staked PHX</div><div class="value">{{.ValidatorSecurity.TotalStaked}}</div></div>
  <div class="metric"><div class="label">Active Validators</div><div class="value">{{.ValidatorSecurity.ActiveValidators}}</div></div>
  <div class="metric"><div class="label">Staking Ratio</div><div class="value">{{.ValidatorSecurity.StakingRatio}}</div></div>
  <div class="metric"><div class="label">Estimated Rewards</div><div class="value">{{.ValidatorSecurity.RewardState}}</div></div>
  <div class="metric"><div class="label">Validator Ranking</div><div class="value">proposal continuity based</div></div>
  <div class="metric"><div class="label">Network Security</div><div class="value">{{.ValidatorSecurity.NetworkHealth}}</div></div>
</section>
<div class="panel">
  <p>Staking is not yet active in the public PhoenixChain devnet. The current network is validator-driven PoA and does not expose delegate, undelegate or rewards flows.</p>
  <p>Canonical monetary base: <strong>{{index .NativeAsset "max_supply_display"}}</strong>.</p>
  <p class="muted">Launch criteria: validator economics, slash conditions, delegation accounting, reward distribution and wallet-safe staking UX.</p>
</div>
{{template "modulePreview" dict "Title" "Current consensus state" "Body" "Live now: PoA devnet with validator list and peer health. Not live yet: staking, rewards, delegations, validator commissions and slash telemetry."}}
<div class="section-title"><h2>Delegation Flow</h2><span class="muted">Professional disabled state</span></div>
<div class="panel">
  <p>1. Select validator</p>
  <p>2. Review health, commission and risk notice</p>
  <p>3. See stake preview</p>
  <p>4. Sign delegation transaction</p>
  <p>5. Verify through explorer receipt</p>
  <p class="muted">Current runtime state: steps 1 and 2 are available as readiness guidance; transaction execution remains disabled until staking backend exists.</p>
</div>
<div class="section-title"><h2>Validator Ranking</h2><span class="muted">Based on live proposal continuity</span></div>
<table><tr><th>Validator</th><th>Health</th><th>Uptime</th><th>Blocks Proposed</th><th>Risk</th><th>Action</th></tr>{{range .ValidatorRecords}}<tr><td>{{.Name}}</td><td><span class="tag">{{.Health}}</span></td><td>{{.UptimeLabel}}</td><td>{{.BlocksProposed}}</td><td>{{.RiskStatus}}</td><td><a href="/validators/{{.Address}}">Review Validator</a></td></tr>{{else}}<tr><td colspan="6" class="muted">No validators available.</td></tr>{{end}}</table>
<div class="section-title"><h2>Reward Visibility</h2><span class="muted">No fake APY</span></div>
<div class="panel">
  <p>Estimated rewards: not published in current devnet runtime.</p>
  <p>Validator rewards: not active in live portal.</p>
  <p>Delegator rewards: not active in live portal.</p>
  <p>Claim rewards: disabled until staking runtime support exists.</p>
</div>
<div class="section-title"><h2>Future Reward Architecture</h2><span class="muted">Placeholder only</span></div>
<table><tr><th>Allocation Bucket</th><th>Placeholder Share</th><th>Current State</th></tr>{{range .SupplyAllocations}}<tr><td>{{.Label}}</td><td>{{.Percentage}}</td><td>{{.Scope}}</td></tr>{{end}}</table>
{{end}}`, data)
}

func (s *Server) handleRPCPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain RPC")
	data["RPCURL"] = "https://chain.phoenixos.us"
	data["ChainID"] = s.genesis.ChainID
	data["WalletSupport"] = "manual address session only"
	data["WalletConnectSupported"] = false
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>RPC + Developers</h1><span class="muted">Public network access and integration guide</span></div>
<section class="grid">
  <div class="metric"><div class="label">RPC Endpoint</div><div class="value hash">{{.RPCURL}}</div></div>
  <div class="metric"><div class="label">Chain ID</div><div class="value">{{.ChainID}}</div></div>
  <div class="metric"><div class="label">Latest Block</div><div class="value">{{.Head.Height}}</div></div>
  <div class="metric"><div class="label">Healthy Peers</div><div class="value">{{.HealthyPeerCount}}</div></div>
  <div class="metric"><div class="label">Indexed Block</div><div class="value">{{.IndexerStatus.IndexedHeight}}</div></div>
  <div class="metric"><div class="label">Index Health</div><div class="value">{{.IndexerStatus.Health}}</div></div>
  <div class="metric"><div class="label">Faucet</div><div class="value">{{.FaucetStatus}}</div></div>
  <div class="metric"><div class="label">Wallet UX</div><div class="value">{{.WalletSupport}}</div></div>
  <div class="metric"><div class="label">Native Asset</div><div class="value">{{index .NativeAsset "max_supply_compact"}}</div></div>
</section>
<section class="hero">
  <div class="panel">
    <h2>Network Config</h2>
    <div class="portal-actions">
      <button class="portal-button ghost" type="button" data-copy="{{.RPCURL}}">Copy RPC</button>
      <button class="portal-button ghost" type="button" data-copy='{{.ChainConfigJSON}}'>Copy Chain Config</button>
      <a class="portal-button secondary" href="/faucet">Open Faucet</a>
      <a class="portal-button secondary" href="/explorer">Open Explorer</a>
      <button class="portal-button secondary" type="button" disabled>Add Network Pending Wallet Adapter</button>
    </div>
    <div class="portal-callout">
      <p>Network name: {{.NetworkName}}</p>
      <p>Chain ID: {{.ChainID}}</p>
      <p>Explorer URL: {{.ExplorerURL}}</p>
      <p>Native currency: {{.NativeCurrencySymbol}}</p>
      <p>Max supply metadata: {{index .NativeAsset "max_supply_display"}}</p>
      <p>Websocket endpoint: <span class="warn">{{if .WebsocketURL}}{{.WebsocketURL}}{{else}}not exposed in current runtime{{end}}</span></p>
    </div>
  </div>
  <div class="panel">
    <h2>Wallet Connection Audit</h2>
    <p class="muted">PhoenixChain does not currently expose MetaMask/EVM wallet compatibility, browser wallet injection, or a custom Phoenix wallet adapter.</p>
    <div class="portal-callout">
      <p>Browser wallet handshake: <span class="portal-bad">not available</span></p>
      <p>Network add flow: <span class="portal-bad">requires future wallet adapter</span></p>
      <p>Signed transaction test flow: <span class="portal-bad">CLI / raw RPC only for now</span></p>
      <p>Explorer-side address tracking: <span class="portal-good">available through manual wallet session</span></p>
    </div>
  </div>
</section>
<div class="section-title"><h2>Core Endpoints</h2><span class="muted">Live right now</span></div>
<div class="panel">
  <p><span class="hash">GET /health</span> runtime health</p>
  <p><span class="hash">GET /api</span> portal API metadata</p>
  <p><span class="hash">GET /chain/head</span> latest chain head</p>
  <p><span class="hash">GET /chain/block/{height}</span> block by height</p>
  <p><span class="hash">GET /account/{address}</span> account state</p>
  <p><span class="hash">POST /tx/send</span> submit signed transaction</p>
  <p><span class="hash">GET /tx/{hash}</span> transaction lookup</p>
  <p><span class="hash">GET /seed/status</span> seed registry status</p>
  <p><span class="hash">GET /seed/peers</span> public seed peers</p>
  <p><span class="hash">GET /metrics</span> Prometheus metrics</p>
  <p><span class="hash">GET /api/phx20/spec</span> PHX-20 standard metadata</p>
  <p><span class="hash">GET /wallet/assets/{address}</span> wallet-visible PHX-20 balances</p>
  <p><span class="hash">GET /api</span> includes native PHX max-supply metadata</p>
  <p><span class="hash">GET /health</span> includes native PHX metadata for clients that need a lightweight chain descriptor</p>
  <hr class="portal-separator">
  <pre class="portal-code">curl https://chain.phoenixos.us/health
curl https://chain.phoenixos.us/chain/head
curl https://chain.phoenixos.us/account/phx...
curl https://chain.phoenixos.us/api/tx/40a2b3...
curl https://chain.phoenixos.us/seed/peers
curl https://chain.phoenixos.us/api/phx20/spec</pre>
  <p class="muted">Websocket endpoint, contract deployment SDK and wallet adapter docs are not yet part of the current runtime.</p>
</div>
<div class="section-title"><h2>Developer Quickstart</h2><span class="muted">Read, inspect and begin integrating</span></div>
<div class="panel">
  <pre class="portal-code"># JavaScript
const head = await fetch('https://chain.phoenixos.us/chain/head').then(r => r.json());
const account = await fetch('https://chain.phoenixos.us/account/phx...').then(r => r.json());

# Python
import requests
head = requests.get('https://chain.phoenixos.us/chain/head').json()
account = requests.get('https://chain.phoenixos.us/account/phx...').json()
spec = requests.get('https://chain.phoenixos.us/api/phx20/spec').json()

# Transaction lookup
curl https://chain.phoenixos.us/api/tx/TX_HASH</pre>
  <p>SDK package: <span class="warn">not published yet</span></p>
  <p>Wallet integration package: <span class="warn">planned, not available</span></p>
</div>
<div class="section-title"><h2>PHX-20 Quickstart</h2><span class="muted">Official fungible token foundation</span></div>
<div class="panel">
  <div class="portal-actions"><a class="portal-button" href="/tokens/factory">Open Token Factory</a><a class="portal-button secondary" href="/tokens">Open Token Registry</a><a class="portal-button ghost" href="/api/phx20/spec">Read Spec JSON</a></div>
  <pre class="portal-code"># PHX-20 foundation
- Standard: PHX-20
- Runtime: PhoenixChain PCVM devnet token foundation
- Required methods: name, symbol, decimals, totalSupply, balanceOf, transfer, approve, allowance, transferFrom
- Required events: Transfer, Approval, Mint, Burn, Paused, Unpaused

# Current public path
1. Deploy through /tokens/factory
2. Registry appears on /tokens
3. Holders + transfer events appear on /tokens/{address}
4. Wallet-visible balances appear through /wallet/assets/{address}</pre>
  <p class="muted">This is a real PHX-20 runtime registry, not fake market data. Public chain transaction calldata plumbing is still maturing.</p>
</div>
<div class="section-title"><h2>PHX Monetary Reference</h2><span class="muted">Canonical native asset metadata</span></div>
<div class="panel">
  <p>Symbol: {{index .NativeAsset "symbol"}} | Decimals: {{index .NativeAsset "decimals"}} | Max supply: {{index .NativeAsset "max_supply_display"}}</p>
  <p>Circulating supply: <span class="warn">{{index .NativeAsset "circulating_note"}}</span></p>
  <p class="muted">Use this value as the canonical PHX reference in wallet adapters, validator economics, governance tooling and future treasury modules.</p>
</div>
<div class="section-title"><h2>Contract Deploy Quickstart</h2><span class="muted">Runtime truth, not fake EVM claims</span></div>
<div class="panel">
  <pre class="portal-code"># Current status
- PCVM/runtime prototypes exist in source
- Public contract deploy CLI/SDK flow is not finalized
- Wallet-signed deploy from portal is not supported yet

# Recommended current path
1. Run PhoenixChain node / devnet locally
2. Use internal VM/contract tooling from source
3. Broadcast signed tx through POST /tx/send when runtime signing flow is finalized

# Planned examples
- deploy contract sample
- call contract sample
- calldata preview
- receipt/log decode
- explorer verification link</pre>
  <p class="muted">No fake npm package, no fake contract write UI, and no false EVM compatibility claim is made here.</p>
</div>
<div class="section-title"><h2>Transaction Test Flow</h2><span class="muted">Professional disabled state</span></div>
<div class="panel">
  <p>Send test transaction from browser: <span class="warn">requires Phoenix wallet adapter support</span></p>
  <p>RPC write method: <span class="warn">available at API level, but browser-signed wallet flow is pending</span></p>
  <p>Use CLI/internal tooling until wallet flow is finalized.</p>
</div>
{{end}}`, data)
}

func (s *Server) handleDevelopersPortal(w http.ResponseWriter, r *http.Request) {
	s.handleRPCPortal(w, r)
}

func (s *Server) handleStatusPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Status")
	data["RPCStatus"] = "healthy"
	data["FaucetStatus"] = "configured"
	if s.faucetKey == nil {
		data["FaucetStatus"] = "not configured"
	}
	data["ContractRuntimeStatus"] = "prototype runtime"
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Status</h1><span class="muted">Runtime visibility</span></div>
<section class="grid">
  <div class="metric"><div class="label">Chain Head</div><div class="value">{{.Head.Height}}</div></div>
  <div class="metric"><div class="label">Indexed Block</div><div class="value">{{.IndexerStatus.IndexedHeight}}</div></div>
  <div class="metric"><div class="label">Peer Count</div><div class="value">{{.PeerCount}}</div></div>
  <div class="metric"><div class="label">Healthy Peers</div><div class="value">{{.HealthyPeerCount}}</div></div>
  <div class="metric"><div class="label">Faucet</div><div class="value">{{.FaucetStatus}}</div></div>
  <div class="metric"><div class="label">Index Lag</div><div class="value">{{.IndexerStatus.IndexingLag}}</div></div>
  <div class="metric"><div class="label">Index Health</div><div class="value">{{.IndexerStatus.Health}}</div></div>
  <div class="metric"><div class="label">PHX-20 Tokens</div><div class="value">{{.PHX20TokenCount}}</div></div>
  <div class="metric"><div class="label">PHX Max Supply</div><div class="value">{{index .NativeAsset "max_supply_compact"}}</div></div>
</section>
<div class="section-title"><h2>Observed Services</h2><span class="muted">Current devnet scope</span></div>
<div class="panel">
  <p>RPC status: <span class="tag">{{.RPCStatus}}</span></p>
  <p>Explorer status: <span class="tag">healthy</span></p>
  <p>Indexer status: <span class="tag">{{.IndexerStatus.Health}}</span></p>
  <p>Latest indexed block: {{.IndexerStatus.IndexedHeight}} / head {{.IndexerStatus.ChainHeadHeight}}</p>
  <p>Indexed blocks: {{.IndexerStatus.BlockCount}} | Indexed transactions: {{.IndexerStatus.TxCount}} | Indexed addresses: {{.IndexerStatus.AddressCount}}</p>
  <p>PHX-20 runtime: <span class="tag">{{index .PHX20State "status"}}</span> | Tokens: {{.PHX20TokenCount}} | Transfer events: {{.PHX20TransferCount}}</p>
  <p>Native asset reference: {{index .NativeAsset "max_supply_display"}} | Circulating: <span class="warn">{{index .NativeAsset "circulating_note"}}</span></p>
  <p>Contract runtime status: <span class="warn">{{.ContractRuntimeStatus}}</span></p>
  <p>Last reindex reason: {{.IndexerStatus.LastIndexedReason}} at {{.IndexerStatus.LastIndexedAt}}</p>
  <p class="muted">This page is intentionally honest about what exists in the live devnet and what still requires indexer and runtime expansion.</p>
</div>
{{end}}`, data)
}

func (s *Server) handleAnalyticsPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Analytics")
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Analytics</h1><span class="muted">Network growth and activity</span></div>
<section class="grid">
  <div class="metric"><div class="label">Total Blocks</div><div class="value">{{.Head.Height}}</div></div>
  <div class="metric"><div class="label">Recent Transactions</div><div class="value">{{len .Transactions}}</div></div>
  <div class="metric"><div class="label">Validator Count</div><div class="value">{{.ValidatorCount}}</div></div>
  <div class="metric"><div class="label">Avg Block Time</div><div class="value">{{.AverageBlockTime}}</div></div>
  <div class="metric"><div class="label">Indexed Addresses</div><div class="value">{{.IndexerStatus.AddressCount}}</div></div>
  <div class="metric"><div class="label">Indexed Contracts</div><div class="value">{{.IndexerStatus.ContractCount}}</div></div>
  <div class="metric"><div class="label">PHX-20 Transfers</div><div class="value">{{.PHX20TransferCount}}</div></div>
  <div class="metric"><div class="label">Index Lag</div><div class="value">{{.IndexerStatus.IndexingLag}}</div></div>
</section>
<div class="panel">
  <p>Analytics are intentionally limited to real runtime signals currently available from the chain and metrics surface.</p>
  <p class="muted">Daily active addresses, fee market analytics and deep time-series charts still require a dedicated historical indexer. PHX-20 registry counts above are real runtime values.</p>
</div>
{{template "modulePreview" dict "Title" "Analytics expansion path" "Body" "Next milestone: blocks/day, tx/day, active addresses, faucet usage trend, validator proposal share, RPC latency history, contract activity and token transfer analytics."}}
{{end}}`, data)
}

func (s *Server) handleBridgePortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Bridge")
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Bridge</h1><span class="muted">Cross-network asset movement</span></div>
<div class="panel">
  <p>PhoenixChain does not yet expose a live bridge. No fake asset bridge UI is shown.</p>
  <p class="muted">Bridge launch requires trust model disclosure, validator or relayer security, asset mapping, rate limits, monitoring and incident response.</p>
</div>
{{template "modulePreview" dict "Title" "Bridge readiness" "Body" "Not live: bridge contracts, relayer network, asset registry, proof verification, bridge explorer. Live now: chain runtime, faucet, explorer, validators and public API."}}
{{end}}`, data)
}

func (s *Server) handleGovernancePortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Governance")
	runtime := s.governanceData()
	metrics := governancePortalMetrics(runtime.Summary)
	records := buildValidatorRecords(s.chain)
	data["GovernanceRuntime"] = runtime
	data["GovernanceMetrics"] = metrics
	data["ValidatorRecords"] = records
	data["ValidatorSecurity"] = validatorSecurity(records)
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Governance</h1><span class="muted">Network control plane</span></div>
<section class="grid">
  <div class="metric"><div class="label">Active Proposals</div><div class="value">{{index .GovernanceMetrics "active"}}</div></div>
  <div class="metric"><div class="label">Passed Proposals</div><div class="value">{{index .GovernanceMetrics "passed"}}</div></div>
  <div class="metric"><div class="label">Rejected Proposals</div><div class="value">{{index .GovernanceMetrics "rejected"}}</div></div>
  <div class="metric"><div class="label">Quorum State</div><div class="value">{{.GovernanceRuntime.Summary.PowerState}}</div></div>
  <div class="metric"><div class="label">Treasury Overview</div><div class="value">{{.GovernanceRuntime.Treasury.GovernanceFunds}}</div></div>
  <div class="metric"><div class="label">Validator Participation</div><div class="value">{{.GovernanceRuntime.Summary.ValidatorActivity}}</div></div>
  <div class="metric"><div class="label">Execution Queue</div><div class="value">{{index .GovernanceMetrics "queued"}} queued | {{index .GovernanceMetrics "blocked"}} blocked</div></div>
  <div class="metric"><div class="label">Governance Health</div><div class="value">{{.GovernanceRuntime.Summary.GovernanceHealth}}</div></div>
</section>
<div class="panel">
  <p>PhoenixChain now carries a persisted governance runtime for proposal lifecycle, audit trail, treasury safety preview and execution-queue posture.</p>
  <p>What still remains blocked is explicit too: live weighted voting, quorum settlement and treasury execution stay disabled until the chain publishes real governance power and execution safety.</p>
  <p>PHX utility scope is broader than gas alone: {{range $index, $item := index .NativeAsset "governance_utilities"}}{{if $index}}, {{end}}{{$item}}{{end}}.</p>
  <p>Canonical supply reference: <strong>{{index .NativeAsset "max_supply_display"}}</strong>.</p>
  <p class="muted">This page is runtime-backed, but it still refuses to invent treasury movement, validator vote power or fake quorum outcomes.</p>
</div>
<div class="section-title"><h2>Proposal Explorer</h2><span class="muted">Governance records</span></div>
<div class="portal-actions" style="margin-bottom:12px"><a class="portal-button" href="/governance/proposals">Open Proposal Explorer</a><a class="portal-button secondary" href="/governance/treasury">Open Treasury Safety</a><a class="portal-button ghost" href="/governance/audit">Open Audit Trail</a></div>
<table><tr><th>ID</th><th>Title</th><th>Type</th><th>Status</th><th>Quorum</th><th>Queue</th><th>Execution</th></tr>{{range .GovernanceRuntime.Proposals}}<tr><td><a href="/governance/proposals/{{.ProposalID}}">{{.ProposalID}}</a></td><td><a href="/governance/proposals/{{.ProposalID}}">{{.Title}}</a></td><td>{{.Type}}</td><td><span class="tag">{{.Status}}</span></td><td>{{.QuorumRequired}}</td><td>{{.QueueState}}</td><td>{{.ExecutionState}}</td></tr>{{else}}<tr><td colspan="7" class="muted">No proposals recorded.</td></tr>{{end}}</table>
<div class="section-title"><h2>Treasury Layer</h2><span class="muted">Governance-aware funds surface</span></div>
<div class="panel">
  <p>Treasury wallet: {{.GovernanceRuntime.Treasury.Wallet}}</p>
  <p>Governance-controlled funds: {{.GovernanceRuntime.Treasury.GovernanceFunds}}</p>
  <p>Ecosystem reserve: {{.GovernanceRuntime.Treasury.EcosystemReserve}}</p>
  <p>Validator reserve: {{.GovernanceRuntime.Treasury.ValidatorReserve}}</p>
  <p>Grant pool: {{.GovernanceRuntime.Treasury.GrantPool}}</p>
  <p>Operational reserve: {{.GovernanceRuntime.Treasury.OperationalReserve}}</p>
  <p>Treasury history: {{.GovernanceRuntime.Treasury.TreasuryHistory}}</p>
</div>
<div class="section-title"><h2>Validator Governance Identity</h2><span class="muted">Participation surface</span></div>
<table><tr><th>Validator</th><th>Governance Identity</th><th>Participation</th><th>Trust</th><th>Stance</th></tr>{{range .ValidatorRecords}}<tr><td><a href="/validators/{{.Address}}">{{.Name}}</a></td><td>{{.GovernanceIdentity}}</td><td>{{.GovernanceParticipation}}</td><td>{{.GovernanceTrust}}</td><td>{{.GovernanceStance}}</td></tr>{{else}}<tr><td colspan="5" class="muted">No validator governance identities available.</td></tr>{{end}}</table>
<div class="section-title"><h2>Community Governance</h2><span class="muted">Participation and discussion layer</span></div>
<div class="panel">
  <p>Community participation: blocked until wallet-linked governance roles and voting power are published.</p>
  <p>Proposal discussion feed: not active in current devnet portal.</p>
  <p>Participation history: visible only through public audit trail right now.</p>
  <p>Latest votes: runtime records remain empty rather than fabricated.</p>
</div>
<div class="section-title"><h2>Governance Security Model</h2><span class="muted">No fake decentralization claims</span></div>
<div class="panel">
  <p>Quorum requirements: proposal-specific, persisted and intentionally blocked while voting power is unpublished.</p>
  <p>Proposal thresholds: proposal-specific and visible on detail pages.</p>
  <p>Emergency controls: operator-managed emergency path remains in effect.</p>
  <p>Execution timelock: queue-visible and attached to treasury safety posture.</p>
  <p>Treasury safety notes:</p>
  <ul class="portal-list">{{range .GovernanceRuntime.Treasury.SafetyNotes}}<li>{{.}}</li>{{end}}</ul>
</div>
<div class="section-title"><h2>Governance Economic Posture</h2><span class="muted">Future voting and treasury layer</span></div>
<table><tr><th>Allocation Bucket</th><th>Placeholder Share</th><th>Why it matters</th></tr>{{range .SupplyAllocations}}<tr><td>{{.Label}}</td><td>{{.Percentage}}</td><td>{{.Scope}}</td></tr>{{end}}</table>
{{end}}`, data)
}

func (s *Server) handleDocsPortal(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Docs")
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Docs</h1><span class="muted">Canonical chain reference surfaces</span></div>
<section class="grid">
  <div class="metric"><div class="label">Native Asset</div><div class="value">{{index .NativeAsset "symbol"}}</div></div>
  <div class="metric"><div class="label">Decimals</div><div class="value">{{index .NativeAsset "decimals"}}</div></div>
  <div class="metric"><div class="label">Max Supply</div><div class="value">{{index .NativeAsset "max_supply_display"}}</div></div>
  <div class="metric"><div class="label">Consensus</div><div class="value">PoA Devnet</div></div>
</section>
<div class="panel">
  <p>Canonical PHX reference for docs and tooling: <strong>{{index .NativeAsset "max_supply_display"}}</strong>.</p>
  <p>Use cases: {{range $index, $item := index .NativeAsset "governance_utilities"}}{{if $index}}, {{end}}{{$item}}{{end}}.</p>
  <p class="muted">Circulating supply, unlock schedule and staking emissions are intentionally omitted until live runtime modules can back them with real state.</p>
</div>
<div class="portal-actions"><a class="portal-button" href="/rpc">Open RPC Docs</a><a class="portal-button secondary" href="/tokenomics">Open Tokenomics</a><a class="portal-button ghost" href="/governance">Open Governance</a></div>
{{end}}`, data)
}

func prefersHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html") || accept == "" || strings.Contains(accept, "*/*")
}

func normalizeAddress(value string) (string, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if len(value) == 43 && strings.HasPrefix(value, "phx") {
		for _, ch := range value[3:] {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
				return "", false
			}
		}
		return value, true
	}
	return "", false
}

func (s *Server) renderBlockDetail(w http.ResponseWriter, block indexer.BlockRecord) {
	data := s.explorerData("PhoenixOS Chain Block")
	data["Block"] = block
	data["BlockTimestamp"] = block.TimestampLabel
	data["Transactions"] = block.Transactions
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Block #{{.Block.Height}}</h1><span class="muted">{{.BlockTimestamp}}</span></div>
<section class="grid">
  <div class="metric"><div class="label">Hash</div><div class="value hash">{{short .Block.Hash}}</div></div>
  <div class="metric"><div class="label">Transactions</div><div class="value">{{.Block.TransactionCount}}</div></div>
  <div class="metric"><div class="label">Validator</div><div class="value hash">{{short .Block.ValidatorAddress}}</div></div>
  <div class="metric"><div class="label">Consensus</div><div class="value">PoA Devnet</div></div>
</section>
<div class="panel">
  <p class="hash">Block hash: {{.Block.Hash}}</p>
  <p class="hash">Parent hash: {{.Block.PreviousHash}}</p>
  <p class="hash">State root: {{.Block.StateRoot}}</p>
  <p class="hash">Transactions root: {{.Block.TxRoot}}</p>
  <p>Receipts root: <span class="warn">not available in current runtime</span></p>
  <p>Gas used / gas limit: <span class="warn">not available in current runtime</span></p>
  <p>Block size: <span class="warn">not available in current runtime</span></p>
  <p>Logs bloom: <span class="warn">not available in current runtime</span></p>
</div>
<div class="section-title"><h2>Included Transactions</h2><span class="muted">Current block payload</span></div>
{{template "txTable" .}}
{{end}}`, data)
}

func (s *Server) renderTransactionDetail(w http.ResponseWriter, transaction indexer.TransactionRecord) {
	data := s.explorerData("PhoenixOS Chain Transaction")
	data["Transaction"] = transaction
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Transaction</h1><span class="muted">Transfer runtime detail</span></div>
<section class="grid">
  <div class="metric"><div class="label">Status</div><div class="value">{{.Transaction.Status}}</div></div>
  <div class="metric"><div class="label">Value</div><div class="value">{{.Transaction.Amount}} PHX</div></div>
  <div class="metric"><div class="label">Nonce</div><div class="value">{{.Transaction.Nonce}}</div></div>
  <div class="metric"><div class="label">Type</div><div class="value">{{.Transaction.Type}}</div></div>
</section>
<div class="panel">
  <p class="hash">Transaction hash: {{.Transaction.Hash}}</p>
  <p class="hash">From: <a href="/address/{{.Transaction.From}}">{{.Transaction.From}}</a></p>
  <p class="hash">To: <a href="/address/{{.Transaction.To}}">{{.Transaction.To}}</a></p>
  <p>Amount: {{.Transaction.Amount}} PHX</p>
  <p>Gas limit: {{.Transaction.GasLimit}}</p>
  <p>Gas price: {{.Transaction.GasPrice}}</p>
  <p>Method: {{.Transaction.Method}}</p>
  <p>Block: <a href="/blocks/{{.Transaction.BlockHeight}}">#{{.Transaction.BlockHeight}}</a></p>
  <p>Execution result: {{.Transaction.Status}}</p>
  <p>Receipt state: <span class="warn">{{if .Transaction.ReceiptAvailable}}indexed{{else}}runtime unavailable{{end}}</span></p>
  <p>Input data / logs / internal calls: <span class="warn">not available in current runtime</span></p>
</div>
{{end}}`, data)
}

func unixTimeLabel(value int64) string {
	if value <= 0 {
		return "genesis runtime marker"
	}
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}
