package rpc

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/indexer"
)

const explorerCSS = `
:root{color-scheme:dark;--bg:#071019;--panel:#0f1724;--panel2:#142034;--text:#edf3fb;--muted:#8fa4bc;--line:#1d3045;--accent:#2de2c5;--accent2:#7cc7ff;--good:#43d39e}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font-family:Inter,ui-sans-serif,system-ui,Segoe UI,Arial,sans-serif}a{color:#7cc7ff;text-decoration:none}a:hover{text-decoration:underline}
header{border-bottom:1px solid var(--line);background:rgba(9,15,24,.92);backdrop-filter:blur(12px);position:sticky;top:0;z-index:2}nav{max-width:1320px;margin:auto;display:flex;align-items:center;justify-content:space-between;padding:16px 24px}.brand{font-weight:800;color:var(--accent);letter-spacing:.3px}.navlinks{display:flex;gap:14px;flex-wrap:wrap}
main{max-width:1320px;margin:auto;padding:24px}.hero{display:grid;grid-template-columns:1.6fr 1fr;gap:18px;margin-bottom:18px}.panel{background:linear-gradient(180deg,rgba(15,23,36,.98),rgba(11,18,30,.98));border:1px solid var(--line);border-radius:14px;padding:18px;box-shadow:0 18px 50px rgba(0,0,0,.28)}.grid{display:grid;grid-template-columns:repeat(4,1fr);gap:14px}.metric{background:var(--panel2);border:1px solid var(--line);border-radius:12px;padding:14px}.metric .label{color:var(--muted);font-size:13px}.metric .value{font-size:24px;font-weight:760;margin-top:6px;overflow-wrap:anywhere}
table{width:100%;border-collapse:collapse;background:var(--panel);border:1px solid var(--line);border-radius:8px;overflow:hidden}th,td{text-align:left;padding:12px;border-bottom:1px solid var(--line);vertical-align:top}th{color:var(--muted);font-size:13px;background:#0f1720}td{font-size:14px}.hash{font-family:ui-monospace,SFMono-Regular,Consolas,monospace;overflow-wrap:anywhere}.tag{display:inline-flex;border:1px solid var(--line);border-radius:999px;padding:3px 8px;color:var(--good);font-size:12px}.section-title{display:flex;justify-content:space-between;align-items:end;margin:22px 0 10px}.muted{color:var(--muted)}.warn{color:#ffc078}.portal-search{display:flex;gap:10px;flex-wrap:wrap;margin-top:16px}.portal-search input,.portal-input{flex:1 1 320px;padding:14px 16px;border-radius:12px;border:1px solid var(--line);background:#0a121d;color:var(--text)}.portal-search button,.portal-button{padding:14px 18px;border-radius:12px;border:0;background:var(--accent);color:#052017;font-weight:800;cursor:pointer}.portal-button.secondary{background:#162436;color:var(--text);border:1px solid var(--line)}.portal-button.ghost{background:transparent;color:var(--accent2);border:1px solid var(--line)}.portal-button[disabled]{opacity:.55;cursor:not-allowed}.portal-actions{display:flex;gap:10px;flex-wrap:wrap}.portal-list{margin:0;padding-left:18px;color:var(--muted)}.portal-code{padding:14px;border-radius:12px;border:1px solid var(--line);background:#0a121d;color:#d9f0ff;overflow:auto}.portal-separator{border:0;border-top:1px solid var(--line);margin:18px 0}.portal-callout{padding:12px 14px;border-radius:12px;border:1px solid var(--line);background:#0c1625;margin-top:12px}.portal-good{color:var(--good)}.portal-bad{color:#ff9d7a}.portal-form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.portal-checkboxes{display:flex;gap:16px;flex-wrap:wrap;margin:16px 0}.asset-list,.asset-list li{list-style:none;padding:0;margin:0}.asset-list li{padding:10px 0;border-bottom:1px solid var(--line)}.asset-empty{color:var(--muted)}
@media(max-width:800px){.hero{grid-template-columns:1fr}.grid{grid-template-columns:1fr 1fr}nav{align-items:flex-start;gap:12px;flex-direction:column}.portal-form-grid{grid-template-columns:1fr}}@media(max-width:520px){.grid{grid-template-columns:1fr}.portal-actions{flex-direction:column}th:nth-child(3),td:nth-child(3){display:none}}
`

var explorerTemplate = template.Must(template.New("explorer").Funcs(template.FuncMap{
	"short": func(value string) string {
		if len(value) <= 18 {
			return value
		}
		return value[:10] + "..." + value[len(value)-8:]
	},
	"dict": func(values ...any) map[string]any {
		out := map[string]any{}
		for i := 0; i+1 < len(values); i += 2 {
			key, _ := values[i].(string)
			out[key] = values[i+1]
		}
		return out
	},
}).Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}}</title><meta http-equiv="refresh" content="12"><style>` + explorerCSS + `</style></head>
<body><header><nav><div class="brand">PhoenixOS Chain</div><div class="navlinks"><a href="/">Home</a><a href="/explorer">Explorer</a><a href="/blocks">Blocks</a><a href="/tx">Transactions</a><a href="/validators">Validators</a><a href="/governance">Governance</a><a href="/contracts">Contracts</a><a href="/tokens">Tokens</a><a href="/tokens/factory">Token Factory</a><a href="/tokenomics">Tokenomics</a><a href="/faucet">Faucet</a><a href="/rpc">RPC</a><a href="/analytics">Analytics</a><a href="/status">Status</a><a href="/docs">Docs</a></div></nav></header>
<main>{{template "content" .}}</main><script>
(() => {
  const storageKey = 'phoenixchain.wallet.session.v1';
  const input = document.querySelector('[data-wallet-input]');
  const connectButton = document.querySelector('[data-wallet-connect]');
  const disconnectButton = document.querySelector('[data-wallet-disconnect]');
  const autofillButton = document.querySelector('[data-wallet-autofill]');
  const faucetInput = document.querySelector('[data-faucet-address]');
  const walletValueNodes = document.querySelectorAll('[data-wallet-address]');
  const walletStateNodes = document.querySelectorAll('[data-wallet-state]');
  const factoryCreatorInput = document.querySelector('[data-wallet-input-factory]');
  const factoryOwnerInput = document.querySelector('[data-wallet-owner-factory]');
  const factoryTreasuryInput = document.querySelector('[data-wallet-treasury-factory]');
  const factoryAutofill = document.querySelector('[data-wallet-factory-autofill]');
  const assetNode = document.querySelector('[data-wallet-assets]');
  const loadSession = () => window.localStorage.getItem(storageKey) || '';
  const loadAssets = async (value) => {
    if (!assetNode) return;
    if (!value) {
      assetNode.textContent = 'Connect or track a Phoenix address to load PHX-20 balances.';
      return;
    }
    assetNode.textContent = 'Loading PHX-20 balances...';
    try {
      const response = await fetch('/wallet/assets/' + encodeURIComponent(value), { headers: { 'Accept': 'application/json' } });
      if (!response.ok) throw new Error('wallet asset fetch failed');
      const payload = await response.json();
      const balances = Array.isArray(payload.balances) ? payload.balances : [];
      if (!balances.length) {
        assetNode.innerHTML = '<div class="asset-empty">No PHX-20 balances indexed for this address yet.</div>';
        return;
      }
      assetNode.innerHTML = '<ul class="asset-list">' + balances.map((item) => '<li><strong>' + item.symbol + '</strong> <span class="muted">' + item.name + '</span><br><span class="hash">' + item.token_address + '</span><br>Balance: ' + item.balance + '</li>').join('') + '</ul>';
    } catch (_) {
      assetNode.textContent = 'Unable to load PHX-20 wallet balances right now.';
    }
  };
  const apply = (value) => {
    walletValueNodes.forEach((node) => node.textContent = value || 'manual session not connected');
    walletStateNodes.forEach((node) => node.textContent = value ? 'manual address session active' : 'browser wallet adapter not connected');
    if (faucetInput && value) faucetInput.value = value;
    if (factoryCreatorInput && !factoryCreatorInput.value) factoryCreatorInput.value = value;
    if (factoryOwnerInput && !factoryOwnerInput.value) factoryOwnerInput.value = value;
    if (factoryTreasuryInput && !factoryTreasuryInput.value) factoryTreasuryInput.value = value;
    loadAssets(value);
  };
  const valid = (value) => /^phx[a-f0-9]{40}$/.test((value || '').toLowerCase());
  const initial = loadSession();
  if (input) input.value = initial;
  apply(initial);
  connectButton?.addEventListener('click', () => {
    const value = (input?.value || '').trim().toLowerCase();
    if (!valid(value)) {
      alert('Enter a valid Phoenix address in phx... format.');
      return;
    }
    window.localStorage.setItem(storageKey, value);
    apply(value);
  });
  disconnectButton?.addEventListener('click', () => {
    window.localStorage.removeItem(storageKey);
    if (input) input.value = '';
    if (factoryCreatorInput) factoryCreatorInput.value = '';
    if (factoryOwnerInput) factoryOwnerInput.value = '';
    if (factoryTreasuryInput) factoryTreasuryInput.value = '';
    apply('');
  });
  autofillButton?.addEventListener('click', () => {
    const value = loadSession();
    if (faucetInput && value) faucetInput.value = value;
  });
  factoryAutofill?.addEventListener('click', () => {
    const value = loadSession();
    if (factoryCreatorInput && value) factoryCreatorInput.value = value;
    if (factoryOwnerInput && value) factoryOwnerInput.value = value;
    if (factoryTreasuryInput && value) factoryTreasuryInput.value = value;
  });
  document.querySelectorAll('[data-copy]').forEach((button) => {
    button.addEventListener('click', async () => {
      const value = button.getAttribute('data-copy');
      if (!value) return;
      try {
        await navigator.clipboard.writeText(value);
        const label = button.textContent;
        button.textContent = 'Copied';
        window.setTimeout(() => { button.textContent = label; }, 1200);
      } catch (_) {}
    });
  });
})();
</script></body></html>`))

func (s *Server) handleExplorerHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
renderExplorer(w, `{{define "content"}}
<section class="hero"><div class="panel"><h1>PhoenixOS Chain Devnet</h1><p class="muted">Live explorer, validator network, faucet, PHX-20 token foundation and public runtime API for the PhoenixOS chain layer. Presented like a serious chain portal, while staying explicit that this is still devnet infrastructure.</p><p><span class="tag">LIVE DEVNET</span> <span class="tag">CHAIN ID {{.ChainID}}</span> <span class="tag">RPC HEALTHY</span> <span class="tag">{{.PHX20TokenCount}} PHX-20 TOKENS</span> <span class="tag">{{index .NativeAsset "max_supply_compact"}}</span></p><form class="portal-search" method="get" action="/search"><input type="text" name="q" placeholder="Search block, tx, address or contract"><button type="submit">Search Network</button></form></div><div class="panel"><h3>Network Command Center</h3><p class="hash">RPC: https://chain.phoenixos.us</p><p class="hash">Chain ID: {{.ChainID}}</p><p>Native asset: {{index .NativeAsset "symbol"}} · {{index .NativeAsset "max_supply_display"}} max supply</p><p>State: devnet, public explorer and API online</p><p>PHX-20 runtime: {{.PHX20State.status}}</p><p class="muted">Mainnet, audited staking and bridge layers are not yet active.</p></div></section>
<section class="hero"><div class="panel"><h2>Wallet Access Panel</h2><p class="muted">PhoenixChain does not yet expose MetaMask or EVM wallet support. You can still track a Phoenix address session locally, autofill faucet requests and inspect live balance/address activity.</p><div class="portal-actions"><input class="portal-input" data-wallet-input placeholder="phx..."><button class="portal-button" type="button" data-wallet-connect>Track Address</button><button class="portal-button secondary" type="button" data-wallet-disconnect>Disconnect</button></div><div class="portal-callout"><p>Status: <span class="portal-good" data-wallet-state>browser wallet adapter not connected</span></p><p class="hash">Tracked address: <span data-wallet-address>manual session not connected</span></p><p>Wallet mode: manual address session only</p><p class="muted">Native browser wallet handshake, add-network button and signed transaction flow remain pending a dedicated Phoenix wallet adapter.</p></div></div><div class="panel"><h2>Network Config</h2><div class="portal-actions"><button class="portal-button ghost" type="button" data-copy="{{.RPCURL}}">Copy RPC</button><button class="portal-button ghost" type="button" data-copy='{{.ChainConfigJSON}}'>Copy Chain Config</button><a class="portal-button secondary" href="/faucet">Open Faucet</a><a class="portal-button secondary" href="/explorer">Open Explorer</a><button class="portal-button secondary" type="button" disabled>Add Network Pending Wallet Adapter</button></div><div class="portal-callout"><p>Network: {{.NetworkName}}</p><p>Explorer: {{.ExplorerURL}}</p><p>Native currency: {{.NativeCurrencySymbol}}</p><p>Websocket: {{if .WebsocketURL}}{{.WebsocketURL}}{{else}}not exposed in current runtime{{end}}</p></div></div></section>
` + walletBalancePanelHTML() + `
<section class="grid"><div class="metric"><div class="label">Latest Block</div><div class="value">{{.Head.Height}}</div></div><div class="metric"><div class="label">Average Block Time</div><div class="value">{{.AverageBlockTime}}</div></div><div class="metric"><div class="label">Total Recent Transactions</div><div class="value">{{len .Transactions}}</div></div><div class="metric"><div class="label">Active Validators</div><div class="value">{{.ValidatorCount}}</div></div><div class="metric"><div class="label">Max Supply</div><div class="value">{{index .NativeAsset "max_supply_compact"}}</div></div><div class="metric"><div class="label">Emission State</div><div class="value">{{index .NativeAsset "emission_state"}}</div></div></section>
<section class="grid"><div class="metric"><div class="label">Mempool</div><div class="value">{{.MempoolSize}}</div></div><div class="metric"><div class="label">Peers</div><div class="value">{{.PeerCount}}</div></div><div class="metric"><div class="label">Healthy Peers</div><div class="value">{{.HealthyPeerCount}}</div></div><div class="metric"><div class="label">Faucet</div><div class="value">{{.FaucetStatus}}</div></div><div class="metric"><div class="label">PHX-20 Tokens</div><div class="value">{{.PHX20TokenCount}}</div></div><div class="metric"><div class="label">PHX-20 Transfers</div><div class="value">{{.PHX20TransferCount}}</div></div></section>
<div class="section-title"><h2>PHX Monetary Architecture</h2><a href="/tokenomics">Open tokenomics</a></div><div class="panel"><p><strong>{{index .NativeAsset "max_supply_display"}}</strong> is the canonical maximum supply reference for PHX across explorer, wallet metadata, RPC docs, governance positioning and future validator economics.</p><p>Circulating supply: <span class="warn">{{index .NativeAsset "circulating_note"}}</span></p><p>Treasury / staking / validator rewards are shown as architecture placeholders only until their live runtime modules activate.</p></div>
<div class="section-title"><h2>Public Testnet</h2><a href="/seed/peers">Seed peers API</a></div><div class="panel"><p>Faucet: <a href="/faucet">request test PHX</a></p><p>Seed status: <a href="/seed/status">/seed/status</a></p><p>Developer hub: <a href="/rpc">/rpc</a></p><p class="muted">External public nodes can register with the seed registry. This network is still test infrastructure, not mainnet.</p></div>
<div class="section-title"><h2>Latest Blocks</h2><a href="/blocks">View all</a></div>{{template "blocksTable" .}}
<div class="section-title"><h2>Latest Transactions</h2><a href="/tx">View all</a></div>{{template "txTable" .}}
{{template "networkReadiness" .}}
{{end}}`, s.explorerData("PhoenixOS Chain Explorer"))
}

func (s *Server) handleExplorerBlocks(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Blocks")
	data["BlocksPage"] = s.indexer.RecentBlocks(pageFromRequest(r), pageSizeFromRequest(r))
	data["Blocks"] = data["BlocksPage"].(indexer.Page[indexer.BlockRecord]).Items
	renderExplorer(w, `{{define "content"}}<div class="section-title"><h1>Latest Blocks</h1><span class="muted">Indexed explorer view</span></div>{{template "blocksTable" .}}{{template "pageState" .BlocksPage}}{{end}}`, data)
}

func (s *Server) handleExplorerTxs(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Transactions")
	data["TxPage"] = s.indexer.RecentTransactions(pageFromRequest(r), pageSizeFromRequest(r))
	data["Transactions"] = data["TxPage"].(indexer.Page[indexer.TransactionRecord]).Items
	renderExplorer(w, `{{define "content"}}<div class="section-title"><h1>Latest Transactions</h1><span class="muted">Confirmed indexed transactions</span></div>{{template "txTable" .}}{{template "pageState" .TxPage}}{{end}}`, data)
}

func (s *Server) handleExplorerValidators(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Validators")
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if sortBy == "" {
		sortBy = "uptime"
	}
	if status == "" {
		status = "all"
	}
	records := filterAndSortValidators(s.governanceAwareValidatorRecords(), query, sortBy, status)
	data["ValidatorQuery"] = query
	data["ValidatorSort"] = sortBy
	data["ValidatorFilter"] = status
	data["ValidatorRecords"] = records
	data["ValidatorSecurity"] = validatorSecurity(records)
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Validators</h1><span class="muted">Validator-secured network surface</span></div>
<div class="panel">
  <p>PhoenixChain currently runs a permissioned PoA validator set. This page shows real validator addresses, proposal continuity and network health derived from live chain/runtime state.</p>
  <p class="muted">Staking, delegation, rewards, slashing and on-chain governance are not faked before their runtime modules exist.</p>
</div>
<section class="grid">
  <div class="metric"><div class="label">Active Validators</div><div class="value">{{.ValidatorSecurity.ActiveValidators}}</div></div>
  <div class="metric"><div class="label">Total Staked PHX</div><div class="value">{{.ValidatorSecurity.TotalStaked}}</div></div>
  <div class="metric"><div class="label">Staking Ratio</div><div class="value">{{.ValidatorSecurity.StakingRatio}}</div></div>
  <div class="metric"><div class="label">Reward State</div><div class="value">{{.ValidatorSecurity.RewardState}}</div></div>
  <div class="metric"><div class="label">Concentration Risk</div><div class="value">{{.ValidatorSecurity.Concentration}}</div></div>
  <div class="metric"><div class="label">Network Health</div><div class="value">{{.ValidatorSecurity.NetworkHealth}}</div></div>
</section>
<form class="portal-actions" method="get" action="/validators" style="margin-bottom:12px">
  <input class="portal-input" name="q" value="{{.ValidatorQuery}}" placeholder="Search validator name, address or identity">
  <select class="portal-input" name="sort" style="max-width:220px">
    <option value="uptime" {{if eq .ValidatorSort "uptime"}}selected{{end}}>Uptime</option>
    <option value="active" {{if eq .ValidatorSort "active"}}selected{{end}}>Health</option>
    <option value="top-stake" {{if eq .ValidatorSort "top-stake"}}selected{{end}}>Top stake</option>
  </select>
  <select class="portal-input" name="status" style="max-width:220px">
    <option value="all" {{if eq .ValidatorFilter "all"}}selected{{end}}>All</option>
    <option value="active" {{if eq .ValidatorFilter "active"}}selected{{end}}>Active</option>
    <option value="delayed" {{if eq .ValidatorFilter "delayed"}}selected{{end}}>Delayed</option>
    <option value="degraded" {{if eq .ValidatorFilter "degraded"}}selected{{end}}>Degraded</option>
    <option value="offline" {{if eq .ValidatorFilter "offline"}}selected{{end}}>Offline</option>
    <option value="unhealthy" {{if eq .ValidatorFilter "unhealthy"}}selected{{end}}>Unhealthy</option>
  </select>
  <button class="portal-button" type="submit">Apply</button>
</form>
<table><tr><th>Name</th><th>Address</th><th>Status</th><th>Uptime</th><th>Blocks Proposed</th><th>Total Stake</th><th>Self Stake</th><th>Delegators</th><th>Commission</th><th>Latest Activity</th><th>Health</th></tr>{{range .ValidatorRecords}}<tr><td><a href="/validators/{{.Address}}">{{.Name}}</a></td><td class="hash"><a href="/validators/{{.Address}}">{{short .Address}}</a></td><td>{{.Status}}</td><td>{{.UptimeLabel}}</td><td>{{.BlocksProposed}}</td><td>{{.TotalStake}}</td><td>{{.SelfStake}}</td><td>{{.Delegators}}</td><td>{{.Commission}}</td><td>{{.LatestActivity}}</td><td><span class="tag">{{.Health}}</span></td></tr>{{else}}<tr><td colspan="11" class="muted">No validators matched this filter.</td></tr>{{end}}</table>
<div class="section-title"><h2>Peer Surface</h2><span class="muted">Network health by reachable peers</span></div>
<table><tr><th>Peer Address</th><th>Healthy</th><th>Failures</th><th>Last error</th></tr>{{range .Peers}}<tr><td class="hash">{{.Address}}</td><td>{{.Healthy}}</td><td>{{.Failures}}</td><td class="muted">{{.LastError}}</td></tr>{{else}}<tr><td colspan="4" class="muted">No peer telemetry recorded yet.</td></tr>{{end}}</table>
{{template "modulePreview" dict "Title" "Validator runtime integrity" "Body" "Live today: validator addresses, proposer continuity, peer health and network security posture. Still pending runtime support: on-chain validator metadata, delegation accounting, staking claims, commissions, slash execution and reward settlement."}}
{{end}}`, data)
}

func (s *Server) handleExplorerBlock(w http.ResponseWriter, r *http.Request) {
	height, err := strconv.ParseUint(r.PathValue("height"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid height")
		return
	}
	block, ok := s.indexer.FindBlockByHeight(height)
	if !ok {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}
	data := s.explorerData("PhoenixOS Chain Block")
	data["Block"] = block
	data["Transactions"] = block.Transactions
	renderExplorer(w, `{{define "content"}}<h1>Block #{{.Block.Height}}</h1><div class="panel"><p class="hash">Hash: {{.Block.Hash}}</p><p class="hash">Parent: {{.Block.PreviousHash}}</p><p>Transactions: {{.Block.TransactionCount}}</p><p class="hash">Validator: {{.Block.ValidatorAddress}}</p></div><h2>Transactions</h2>{{template "txTable" .}}{{end}}`, data)
}

func (s *Server) handleExplorerTx(w http.ResponseWriter, r *http.Request) {
	transaction, ok := s.indexer.FindTransaction(r.PathValue("hash"))
	if !ok {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	data := s.explorerData("PhoenixOS Chain Transaction")
	data["Transaction"] = transaction
	renderExplorer(w, `{{define "content"}}<h1>Transaction</h1><div class="panel"><p class="hash">Hash: {{.Transaction.Hash}}</p><p class="hash">From: <a href="/address/{{.Transaction.From}}">{{.Transaction.From}}</a></p><p class="hash">To: <a href="/address/{{.Transaction.To}}">{{.Transaction.To}}</a></p><p>Amount: {{.Transaction.Amount}} PHX</p><p>Nonce: {{.Transaction.Nonce}}</p><p>Status: {{.Transaction.Status}}</p></div>{{end}}`, data)
}

func (s *Server) handleExplorerAccount(w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")
	data := s.explorerData("PhoenixOS Chain Account")
	record, txPage, ok := s.indexer.AddressActivity(address, pageFromRequest(r), pageSizeFromRequest(r))
	if !ok {
		writeError(w, http.StatusNotFound, "account not indexed")
		return
	}
	data["Account"] = record
	data["Transactions"] = txPage.Items
	data["TxPage"] = txPage
	renderExplorer(w, `{{define "content"}}<h1>Account</h1><div class="panel"><p class="hash">Address: {{.Account.Address}}</p><p>Balance: {{.Account.Balance}} PHX</p><p>Nonce: {{.Account.Nonce}}</p><p>Tx Count: {{.Account.ActivityCounter.Transactions}}</p></div><h2>Recent Activity</h2>{{template "txTable" .}}{{template "pageState" .TxPage}}{{end}}`, data)
}

func (s *Server) explorerData(title string) map[string]any {
	status := s.indexer.Status()
	blocksPage := s.indexer.RecentBlocks(1, 20)
	txPage := s.indexer.RecentTransactions(1, 25)
	maxSupplyDisplay := fmt.Sprintf("%s %s", formatWholeNumber(s.genesis.NativeCurrency.MaxSupply), s.genesis.NativeCurrency.Symbol)
	maxSupplyCompact := fmt.Sprintf("%dM %s", s.genesis.NativeCurrency.MaxSupply/1000000, s.genesis.NativeCurrency.Symbol)
	chainConfigJSON := fmt.Sprintf(`{"chainName":"PhoenixChain Devnet","chainId":%d,"rpcUrl":"https://chain.phoenixos.us","explorerUrl":"https://chain.phoenixos.us","nativeCurrency":{"name":"Phoenix","symbol":"PHX","decimals":18},"nativeAssetMetadata":{"symbol":"PHX","decimals":18,"maxSupply":%d,"maxSupplyDisplay":"%s","maxSupplyCompact":"%s"}}`, s.genesis.ChainID, s.genesis.NativeCurrency.MaxSupply, maxSupplyDisplay, maxSupplyCompact)
	return map[string]any{
		"Title":            title,
		"Head":             s.chain.Head(),
		"Blocks":           blocksPage.Items,
		"Transactions":     txPage.Items,
		"MempoolSize":      len(s.mempool.All()),
		"PeerCount":        len(s.p2p.SnapshotPeers()),
		"HealthyPeerCount": s.p2p.HealthyPeerCount(),
		"Peers":            s.p2p.SnapshotPeerStatuses(),
		"ValidatorCount":   len(s.chain.ValidatorList()),
		"Validators":       s.chain.ValidatorList(),
		"AverageBlockTime": averageBlockTime(blocksPage.Items),
		"ChainID":          s.genesis.ChainID,
		"FaucetStatus":     faucetStateLabel(s.faucetKey != nil),
		"IndexerStatus":    status,
		"PHX20State":       s.phx20.TokenState(),
		"PHX20TokenCount":  len(s.phx20.ListTokens()),
		"PHX20TransferCount": s.phx20.TokenState()["indexed_transfer_count"],
		"RPCURL":           "https://chain.phoenixos.us",
		"ExplorerURL":      "https://chain.phoenixos.us",
		"NetworkName":      "PhoenixChain Devnet",
		"NativeCurrencySymbol": "PHX",
		"NativeAsset":      nativeAssetMetadata(s.genesis),
		"SupplyAllocations": supplyAllocations(),
		"WebsocketURL":     "",
		"ChainConfigJSON":  chainConfigJSON,
	}
}

func pageFromRequest(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		return 1
	}
	return page
}

func pageSizeFromRequest(r *http.Request) int {
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize <= 0 {
		return indexer.DefaultPageSize
	}
	if pageSize > indexer.MaxPageSize {
		return indexer.MaxPageSize
	}
	return pageSize
}

func renderExplorer(w http.ResponseWriter, content string, data map[string]any) {
	tpl := template.Must(explorerTemplate.Clone())
	template.Must(tpl.Parse(`{{define "blocksTable"}}<table><tr><th>Height</th><th>Hash</th><th>Txs</th><th>Validator</th></tr>{{range .Blocks}}<tr><td><a href="/blocks/{{.Height}}">#{{.Height}}</a></td><td class="hash">{{short .Hash}}</td><td>{{.TransactionCount}}</td><td class="hash">{{short .ValidatorAddress}}</td></tr>{{else}}<tr><td colspan="4" class="muted">No indexed blocks yet.</td></tr>{{end}}</table>{{end}}{{define "txTable"}}<table><tr><th>Hash</th><th>From</th><th>To</th><th>Amount</th></tr>{{range .Transactions}}<tr><td class="hash"><a href="/tx/{{.Hash}}">{{short .Hash}}</a></td><td class="hash"><a href="/address/{{.From}}">{{short .From}}</a></td><td class="hash"><a href="/address/{{.To}}">{{short .To}}</a></td><td>{{.Amount}}</td></tr>{{else}}<tr><td colspan="4" class="muted">No indexed transactions yet.</td></tr>{{end}}</table>{{end}}{{define "modulePreview"}}<div class="section-title"><h2>{{.Title}}</h2><span class="muted">Professional preview state</span></div><div class="panel"><p>{{.Body}}</p></div>{{end}}{{define "networkReadiness"}}<div class="section-title"><h2>Runtime readiness</h2><span class="muted">Truthful current state</span></div><div class="panel"><p>PhoenixChain is live here as a devnet explorer and public API surface. It is not mainnet, not audited, not fully decentralized and not yet backed by a complete explorer indexer.</p><p>Indexed height: {{.IndexerStatus.IndexedHeight}} | Chain head: {{.IndexerStatus.ChainHeadHeight}} | Lag: {{.IndexerStatus.IndexingLag}} | Health: {{.IndexerStatus.Health}}</p></div>{{end}}{{define "pageState"}}<div class="panel" style="margin-top:12px"><p class="muted">Page {{.Page}} / {{.TotalPages}} · {{.TotalItems}} indexed items</p></div>{{end}}`))
	template.Must(tpl.Parse(content))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, data)
}

func averageBlockTime(blocks []indexer.BlockRecord) string {
	if len(blocks) < 2 {
		return "not enough data"
	}
	var last int64
	var total int64
	var count int64
	for _, block := range blocks {
		if block.Timestamp <= 0 {
			continue
		}
		if last > 0 && last > block.Timestamp {
			total += last - block.Timestamp
			count++
		}
		last = block.Timestamp
	}
	if count == 0 {
		return "not enough data"
	}
	return fmt.Sprintf("%ds", total/count)
}

func faucetStateLabel(configured bool) string {
	if configured {
		return "configured"
	}
	return "not configured"
}
