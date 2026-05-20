package rpc

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/phx20"
)

func (s *Server) handlePHX20Spec(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, phx20.CurrentSpec())
}

func (s *Server) handleWalletAssets(w http.ResponseWriter, r *http.Request) {
	address, ok := normalizeAddress(r.PathValue("address"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid address")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"address":      address,
		"balances":     s.phx20.AddressBalances(address),
		"native_asset": nativeAssetMetadata(s.genesis),
	})
}

func (s *Server) handleTokenFactoryPage(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Token Factory")
	data["Templates"] = phx20.Templates()
	data["Spec"] = phx20.CurrentSpec()
	data["RecommendedDecimals"] = 18
	data["PreviewTemplate"] = phx20.Templates()[0]
	data["FactoryStatus"] = strings.TrimSpace(r.URL.Query().Get("status"))
	data["FactoryMessage"] = strings.TrimSpace(r.URL.Query().Get("message"))
	data["FactoryTokenAddress"] = strings.TrimSpace(r.URL.Query().Get("token"))
	data["FactoryRiskWarnings"] = []string{
		"Browser deploy is devnet-only and unaudited.",
		"Wallet-signed deploy flow is not live yet; creator/owner/treasury validation is enforced server-side.",
		"Logo and metadata URLs must be safe http/https links.",
	}
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>PHX-20 Token Factory</h1><span class="muted">Official fungible token deploy surface for PhoenixChain devnet</span></div>
<section class="hero">
  <div class="panel">
    <h2>Deploy a PHX-20 Token</h2>
    <p class="muted">This factory performs a real PCVM-backed devnet token deploy into the PhoenixChain PHX-20 registry. It does not pretend to be mainnet and it does not fake wallet-signing support.</p>
    {{if .FactoryStatus}}<div class="portal-callout"><p>Status: <strong>{{.FactoryStatus}}</strong></p><p>{{.FactoryMessage}}</p>{{if .FactoryTokenAddress}}<p class="hash">Token: <a href="/tokens/{{.FactoryTokenAddress}}">{{.FactoryTokenAddress}}</a></p>{{end}}</div>{{end}}
    <form method="post" action="/tokens/factory/deploy">
      <div class="portal-form-grid">
        <input class="portal-input" name="name" placeholder="Token name" required maxlength="48">
        <input class="portal-input" name="symbol" placeholder="Symbol" required maxlength="12">
        <input class="portal-input" name="supply" type="number" min="1" placeholder="Initial supply" required>
        <input class="portal-input" name="decimals" type="number" min="0" max="18" value="18" required>
        <input class="portal-input" name="creator" data-wallet-input-factory placeholder="Creator phx..." required>
        <input class="portal-input" name="owner" data-wallet-owner-factory placeholder="Owner phx..." required>
        <input class="portal-input" name="treasury" data-wallet-treasury-factory placeholder="Treasury phx..." required>
        <select class="portal-input" name="template">
          {{range .Templates}}<option value="{{.ID}}">{{.Name}}</option>{{end}}
        </select>
        <input class="portal-input" name="logo_url" placeholder="Logo URL (https://...)">
        <input class="portal-input" name="website_url" placeholder="Website URL (https://...)">
        <input class="portal-input" name="twitter_url" placeholder="Social URL (https://x.com/... or similar)">
        <input class="portal-input" name="github_url" placeholder="GitHub URL (https://github.com/...)">
      </div>
      <textarea class="portal-input" name="description" placeholder="Description" maxlength="280" style="min-height:104px;margin-top:12px"></textarea>
      <div class="portal-checkboxes">
        <label><input type="checkbox" name="mintable" value="1"> Mintable</label>
        <label><input type="checkbox" name="burnable" value="1"> Burnable</label>
        <label><input type="checkbox" name="pausable" value="1"> Pausable</label>
        <label><input type="checkbox" name="governance_ready" value="1"> Governance-ready</label>
      </div>
      <div class="portal-actions">
        <button class="portal-button" type="submit">Deploy PHX-20</button>
        <button class="portal-button secondary" type="button" data-wallet-factory-autofill>Use tracked wallet</button>
        <a class="portal-button ghost" href="/tokens">Open Token Registry</a>
      </div>
    </form>
  </div>
  <div class="panel">
    <h2>Deploy Preview</h2>
    <p>Standard: <span class="tag">PHX-20</span></p>
    <p>Runtime: {{.Spec.Runtime}}</p>
    <p>Selected template default: {{.PreviewTemplate.Name}}</p>
    <p>Permission summary: owner + treasury controlled, with optional mint/burn/pause hooks</p>
    <p>Verification status: contract deploys into live registry, source verification still pending</p>
    <p>Required methods: {{range $index, $item := .Spec.RequiredFields}}{{if $index}}, {{end}}{{$item}}{{end}}</p>
    <p>Required events: {{range $index, $item := .Spec.RequiredEvents}}{{if $index}}, {{end}}{{$item}}{{end}}</p>
    <p>Integrity notes:</p>
    <ul class="portal-list">{{range .Spec.IntegrityNotes}}<li>{{.}}</li>{{end}}</ul>
    <div class="portal-callout">
      <p>Gas estimate: template-dependent, exposed per template below</p>
      <p>Audit status: unaudited devnet standard foundation</p>
      <p>Risk warnings:</p>
      <ul class="portal-list">{{range .FactoryRiskWarnings}}<li>{{.}}</li>{{end}}</ul>
    </div>
  </div>
</section>
<div class="section-title"><h2>Templates</h2><span class="muted">Standardized PHX-20 contract presets</span></div>
<table><tr><th>Template</th><th>Capabilities</th><th>Gas</th><th>Description</th></tr>{{range .Templates}}<tr><td>{{.Name}}</td><td>{{if .Mintable}}mint {{end}}{{if .Burnable}}burn {{end}}{{if .Pausable}}pause {{end}}{{if .Governance}}governance{{end}}{{if and (not .Mintable) (not .Burnable) (not .Pausable) (not .Governance)}}basic{{end}}</td><td>{{.GasEstimate}}</td><td>{{.Description}}</td></tr>{{end}}</table>
{{end}}`, data)
}

func (s *Server) handleTokenFactoryDeploy(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/tokens/factory?status=error&message=form+parse+failed", http.StatusSeeOther)
		return
	}
	supply, err := strconv.ParseUint(strings.TrimSpace(r.FormValue("supply")), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/tokens/factory?status=error&message=invalid+supply", http.StatusSeeOther)
		return
	}
	decimals64, err := strconv.ParseUint(strings.TrimSpace(r.FormValue("decimals")), 10, 8)
	if err != nil {
		http.Redirect(w, r, "/tokens/factory?status=error&message=invalid+decimals", http.StatusSeeOther)
		return
	}
	config := phx20.DeployConfig{
		Creator:  strings.ToLower(strings.TrimSpace(r.FormValue("creator"))),
		Owner:    strings.ToLower(strings.TrimSpace(r.FormValue("owner"))),
		Treasury: strings.ToLower(strings.TrimSpace(r.FormValue("treasury"))),
		Name:     strings.TrimSpace(r.FormValue("name")),
		Symbol:   strings.ToUpper(strings.TrimSpace(r.FormValue("symbol"))),
		Decimals: uint8(decimals64),
		Supply:   supply,
		Template: strings.TrimSpace(strings.ToLower(r.FormValue("template"))),
		Mintable: phx20.ParseBoolFlag(r.FormValue("mintable")) || phx20.ParseBoolFlag(r.FormValue("governance_ready")),
		Burnable: phx20.ParseBoolFlag(r.FormValue("burnable")) || phx20.ParseBoolFlag(r.FormValue("governance_ready")),
		Pausable: phx20.ParseBoolFlag(r.FormValue("pausable")) || phx20.ParseBoolFlag(r.FormValue("governance_ready")),
		LogoURL: strings.TrimSpace(r.FormValue("logo_url")),
		Description: strings.TrimSpace(r.FormValue("description")),
		WebsiteURL: strings.TrimSpace(r.FormValue("website_url")),
		TwitterURL: strings.TrimSpace(r.FormValue("twitter_url")),
		GithubURL: strings.TrimSpace(r.FormValue("github_url")),
	}
	token, _, err := s.phx20.Deploy(config)
	if err != nil {
		http.Redirect(w, r, "/tokens/factory?status=error&message="+urlQuery(err.Error()), http.StatusSeeOther)
		return
	}
	if err := s.save(); err != nil {
		http.Redirect(w, r, "/tokens/factory?status=error&message="+urlQuery("deploy succeeded but snapshot save failed: "+err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/tokens/"+token.Address+"?status=deployed&message="+urlQuery("PHX-20 token deployed into devnet registry")+"&tx="+token.DeployTxHash, http.StatusSeeOther)
}

func urlQuery(value string) string {
	replacer := strings.NewReplacer(" ", "+", "&", "%26", "?", "%3F", "#", "%23")
	return replacer.Replace(value)
}

func walletBalancePanelHTML() string {
	return `<div class="section-title"><h2>Tracked Wallet Token Balances</h2><span class="muted">Manual session asset view</span></div><div class="panel" data-wallet-assets>Connect or track a Phoenix address to load PHX-20 balances.</div>`
}
