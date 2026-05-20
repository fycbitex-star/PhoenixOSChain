package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/tx"
)

const faucetAmount uint64 = 1000
const faucetCooldown = 10 * time.Minute

type faucetRequest struct {
	Address string `json:"address"`
}

type faucetGrantRecord struct {
	Address        string `json:"address"`
	Hash           string `json:"hash"`
	Amount         uint64 `json:"amount"`
	RequestedLabel string `json:"requested_label"`
	Status         string `json:"status"`
}

func (s *Server) handleFaucetPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/faucet" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	data := s.explorerData("PhoenixOS Chain Faucet")
	data["FaucetAddress"] = s.faucetAddr
	data["FaucetAmount"] = faucetAmount
	data["FaucetBalance"] = s.chain.State().GetAccount(s.faucetAddr).Balance
	data["RecentFaucetRequests"] = s.recentFaucetRequests(10)
	data["CooldownMinutes"] = int(faucetCooldown / time.Minute)
	data["FaucetStatusMessage"] = r.URL.Query().Get("message")
	data["FaucetStatus"] = r.URL.Query().Get("status")
	data["FaucetRequestHash"] = r.URL.Query().Get("hash")
	data["FaucetRequestAddress"] = r.URL.Query().Get("address")
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Testnet Faucet</h1><span class="muted">Production-grade devnet token access</span></div>
<section class="grid">
  <div class="metric"><div class="label">Faucet Balance</div><div class="value">{{.FaucetBalance}}</div></div>
  <div class="metric"><div class="label">Latest Block</div><div class="value">{{.Head.Height}}</div></div>
  <div class="metric"><div class="label">Indexed Block</div><div class="value">{{.IndexerStatus.IndexedHeight}}</div></div>
  <div class="metric"><div class="label">Index Health</div><div class="value">{{.IndexerStatus.Health}}</div></div>
</section>
<section class="hero">
<div class="panel">
  <h2>Request Test PHX</h2>
  <p class="muted">Request test PHX for the public PhoenixChain devnet. This faucet does not fake success: acceptance is only shown when a transaction hash is produced.</p>
  {{if .FaucetStatus}}<div class="portal-callout"><p>Status: <strong>{{.FaucetStatus}}</strong></p>{{if .FaucetStatusMessage}}<p>{{.FaucetStatusMessage}}</p>{{end}}{{if .FaucetRequestHash}}<p class="hash">Transaction: <a href="/tx/{{.FaucetRequestHash}}">{{.FaucetRequestHash}}</a></p>{{end}}{{if .FaucetRequestAddress}}<p class="hash">Requested address: {{.FaucetRequestAddress}}</p>{{end}}</div>{{end}}
  <form method="post" action="/faucet/request" class="portal-search">
    <input class="portal-input" data-faucet-address name="address" placeholder="phx...">
    <button class="portal-button" type="submit">Request {{.FaucetAmount}} PHX</button>
    <button class="portal-button secondary" type="button" data-wallet-autofill>Autofill Tracked Address</button>
  </form>
  <div class="portal-callout">
    <p>Cooldown: one request per IP every {{.CooldownMinutes}} minutes</p>
    <p>Daily limit: not configured in current runtime</p>
    <p>Abuse protection: IP cooldown active, stricter captcha/rate controls pending</p>
    <p>Faucet balance: {{.FaucetBalance}} PHX</p>
  </div>
</div>
<div class="panel">
  <h2>Wallet Session</h2>
  <p class="muted">PhoenixChain does not yet expose MetaMask or EVM wallet support. Use the manual wallet session to track an address and feed it into the faucet flow.</p>
  <div class="portal-actions">
    <input class="portal-input" data-wallet-input placeholder="phx...">
    <button class="portal-button" type="button" data-wallet-connect>Track Address</button>
    <button class="portal-button secondary" type="button" data-wallet-disconnect>Disconnect</button>
  </div>
  <div class="portal-callout">
    <p>Status: <span class="portal-good" data-wallet-state>browser wallet adapter not connected</span></p>
    <p class="hash">Tracked address: <span data-wallet-address>manual session not connected</span></p>
    <p>Network: {{.NetworkName}} | Chain ID: {{.ChainID}}</p>
    <div class="portal-actions"><button class="portal-button ghost" type="button" data-copy="{{.RPCURL}}">Copy RPC</button><button class="portal-button secondary" type="button" disabled>Send Test TX Pending Wallet Adapter</button></div>
  </div>
</div>
</section>
<div class="section-title"><h2>Recent Faucet Requests</h2><span class="muted">Accepted faucet grants only</span></div>
<table><tr><th>Address</th><th>Amount</th><th>Status</th><th>Transaction</th><th>Requested</th></tr>{{range .RecentFaucetRequests}}<tr><td class="hash"><a href="/address/{{.Address}}">{{short .Address}}</a></td><td>{{.Amount}} PHX</td><td>{{.Status}}</td><td class="hash"><a href="/tx/{{.Hash}}">{{short .Hash}}</a></td><td>{{.RequestedLabel}}</td></tr>{{else}}<tr><td colspan="5" class="muted">No faucet grants recorded yet.</td></tr>{{end}}</table>
{{end}}`, data)
}

func (s *Server) handleFaucetRequest(w http.ResponseWriter, r *http.Request) {
	respondError := func(status int, message string) {
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") || strings.Contains(r.Header.Get("Accept"), "application/json") {
			writeError(w, status, message)
			return
		}
		http.Redirect(w, r, "/faucet?status=error&message="+url.QueryEscape(message), http.StatusSeeOther)
	}
	if s.faucetKey == nil {
		respondError(http.StatusServiceUnavailable, "faucet is not configured")
		return
	}
	var address string
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req faucetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(http.StatusBadRequest, "invalid faucet request")
			return
		}
		address = req.Address
	} else {
		if err := r.ParseForm(); err != nil {
			respondError(http.StatusBadRequest, "invalid form")
			return
		}
		address = r.FormValue("address")
	}
	if !validPHXAddress(address) {
		respondError(http.StatusBadRequest, "invalid address")
		return
	}
	ip := clientIP(r)
	s.faucetMu.Lock()
	if last, ok := s.faucetLast[ip]; ok && time.Since(last) < faucetCooldown {
		s.faucetMu.Unlock()
		respondError(http.StatusTooManyRequests, "faucet cooldown active")
		return
	}
	s.faucetLast[ip] = time.Now()
	s.faucetMu.Unlock()

	account := s.chain.State().GetAccount(s.faucetAddr)
	transaction := tx.New(s.faucetAddr, address, faucetAmount, account.Nonce, 21000, 1)
	if err := tx.SignTransaction(&transaction, s.faucetKey); err != nil {
		respondError(http.StatusInternalServerError, "faucet signing failed")
		return
	}
	if err := s.mempool.Add(transaction); err != nil {
		respondError(http.StatusBadRequest, fmt.Sprintf("faucet tx rejected: %s", err.Error()))
		return
	}
	s.metrics.IncFaucetRequest()
	s.metrics.IncTxAccepted()
	s.p2p.BroadcastTransaction(transaction)
	s.recordFaucetGrant(address, transaction.Hash)
	if strings.Contains(r.Header.Get("Accept"), "application/json") || strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		writeJSON(w, http.StatusAccepted, map[string]any{"hash": transaction.Hash, "amount": faucetAmount, "address": address})
		return
	}
	http.Redirect(w, r, "/faucet?status=accepted&hash="+url.QueryEscape(transaction.Hash)+"&address="+url.QueryEscape(address)+"&message="+url.QueryEscape("Faucet request accepted and broadcast to the devnet mempool."), http.StatusSeeOther)
}

func validPHXAddress(address string) bool {
	if len(address) != 43 || address[:3] != "phx" {
		return false
	}
	for _, ch := range address[3:] {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

func (s *Server) recordFaucetGrant(address, hash string) {
	s.faucetMu.Lock()
	defer s.faucetMu.Unlock()
	record := faucetGrantRecord{
		Address:        address,
		Hash:           hash,
		Amount:         faucetAmount,
		RequestedLabel: time.Now().UTC().Format(time.RFC3339),
		Status:         "accepted",
	}
	s.faucetHistory = append([]faucetGrantRecord{record}, s.faucetHistory...)
	if len(s.faucetHistory) > 20 {
		s.faucetHistory = s.faucetHistory[:20]
	}
}

func (s *Server) recentFaucetRequests(limit int) []faucetGrantRecord {
	s.faucetMu.Lock()
	defer s.faucetMu.Unlock()
	if limit <= 0 || limit > len(s.faucetHistory) {
		limit = len(s.faucetHistory)
	}
	out := make([]faucetGrantRecord, limit)
	copy(out, s.faucetHistory[:limit])
	return out
}
