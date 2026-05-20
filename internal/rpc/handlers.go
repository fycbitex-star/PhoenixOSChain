package rpc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/p2p"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleExplorerHome)
	mux.HandleFunc("GET /explorer", s.handleExplorerAlias)
	mux.HandleFunc("GET /search", s.handleSearch)
	mux.HandleFunc("GET /blocks", s.handleExplorerBlocks)
	mux.HandleFunc("GET /blocks/{id}", s.handleBlockRoute)
	mux.HandleFunc("GET /tx", s.handleTxListAlias)
	mux.HandleFunc("GET /txs", s.handleExplorerTxs)
	mux.HandleFunc("GET /tx/{hash}", s.handleTxRoute)
	mux.HandleFunc("GET /validators", s.handleExplorerValidators)
	mux.HandleFunc("GET /validators/{address}", s.handleValidatorDetail)
	mux.HandleFunc("GET /address/{address}", s.handleAddressPage)
	mux.HandleFunc("GET /contracts", s.handleContractsIndex)
	mux.HandleFunc("GET /contracts/{address}", s.handleContractDetail)
	mux.HandleFunc("GET /tokens", s.handleTokensIndex)
	mux.HandleFunc("GET /tokens/{address}", s.handleTokenDetail)
	mux.HandleFunc("GET /tokens/factory", s.handleTokenFactoryPage)
	mux.HandleFunc("GET /token-factory", s.handleTokenFactoryPage)
	mux.HandleFunc("GET /tokens/create", s.handleTokenFactoryPage)
	mux.HandleFunc("GET /developers/token-factory", s.handleTokenFactoryPage)
	mux.HandleFunc("GET /tokenomics", s.handleTokenomicsPortal)
	mux.HandleFunc("GET /staking", s.handleStakingPortal)
	mux.HandleFunc("GET /faucet", s.handleFaucetPage)
	mux.HandleFunc("GET /rpc", s.handleRPCPortal)
	mux.HandleFunc("GET /developers", s.handleDevelopersPortal)
	mux.HandleFunc("GET /analytics", s.handleAnalyticsPortal)
	mux.HandleFunc("GET /status", s.handleStatusPortal)
	mux.HandleFunc("GET /bridge", s.handleBridgePortal)
	mux.HandleFunc("GET /governance", s.handleGovernancePortal)
	mux.HandleFunc("GET /governance/proposals", s.handleGovernanceProposals)
	mux.HandleFunc("GET /governance/proposals/{proposalId}", s.handleGovernanceProposalDetail)
	mux.HandleFunc("GET /governance/treasury", s.handleGovernanceTreasury)
	mux.HandleFunc("GET /governance/audit", s.handleGovernanceAudit)
	mux.HandleFunc("GET /docs", s.handleDocsPortal)
	mux.HandleFunc("POST /faucet/request", s.handleFaucetRequest)
	mux.HandleFunc("POST /tokens/factory/deploy", s.handleTokenFactoryDeploy)
	mux.HandleFunc("GET /block-view/{height}", s.handleExplorerBlock)
	mux.HandleFunc("GET /tx-view/{hash}", s.handleExplorerTx)
	mux.HandleFunc("GET /account-view/{address}", s.handleExplorerAccount)
	mux.HandleFunc("GET /wallet/assets/{address}", s.handleWalletAssets)
	mux.HandleFunc("GET /api/phx20/spec", s.handlePHX20Spec)
	mux.HandleFunc("GET /api", s.handleRootAPI)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /chain/head", s.handleHead)
	mux.HandleFunc("GET /chain/latest", s.handleHead)
	mux.HandleFunc("GET /chain/block/{height}", s.handleBlock)
	mux.HandleFunc("GET /account/{address}", s.handleAccount)
	mux.HandleFunc("GET /api/tx/{hash}", s.handleGetTx)
	mux.HandleFunc("POST /tx/send", s.handleSendTx)
	mux.HandleFunc("GET /mempool", s.handleMempool)
	mux.HandleFunc("GET /peers", s.handlePeers)
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	mux.HandleFunc("GET /seed/status", s.handleSeedStatus)
	mux.HandleFunc("GET /seed/peers", s.handleSeedPeers)
	mux.HandleFunc("POST /seed/register", s.handleSeedRegister)

	mux.HandleFunc("GET /p2p/status", s.handleP2PStatus)
	mux.HandleFunc("GET /p2p/block/{height}", s.handleP2PBlock)
	mux.HandleFunc("POST /p2p/tx", s.handleP2PTx)
	mux.HandleFunc("POST /p2p/block", s.handleP2PBlockMessage)
}

func (s *Server) handleRootAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "PhoenixOS Chain Devnet",
		"version":     "v0.8-phx20-foundation",
		"network":     "phoenixos-chain-devnet",
		"production":  false,
		"mainnet":     false,
		"health":      "/health",
		"chain_head":  "/chain/head",
		"mempool":     "/mempool",
		"phx20_spec":  "/api/phx20/spec",
		"native_asset": nativeAssetMetadata(s.genesis),
		"description": "PhoenixOS Chain devnet API with explorer, PHX-20 token foundation, faucet, validator and JSON endpoints. Not production mainnet.",
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := s.indexer.Status()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"height":         s.chain.Height(),
		"indexed_height": status.IndexedHeight,
		"index_lag":      status.IndexingLag,
		"index_health":   status.Health,
		"native_asset":   nativeAssetMetadata(s.genesis),
	})
}

func (s *Server) handleHead(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.chain.Head())
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	height, err := strconv.ParseUint(r.PathValue("height"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid height")
		return
	}
	block, ok := s.chain.GetBlock(height)
	if !ok {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}
	writeJSON(w, http.StatusOK, block)
}

func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.chain.State().GetAccount(r.PathValue("address")))
}

func (s *Server) handleSendTx(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "" && !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}
	var transaction tx.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, "invalid transaction")
		return
	}
	if err := s.mempool.Add(transaction); err != nil {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.metrics.IncTxAccepted()
	logEvent("tx.accepted", map[string]any{"hash": transaction.Hash, "from": transaction.From, "to": transaction.To, "amount": transaction.Amount})
	s.p2p.BroadcastTransaction(transaction)
	writeJSON(w, http.StatusAccepted, map[string]string{"hash": transaction.Hash})
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.p2p.SnapshotPeerStatuses())
}

func (s *Server) handleGetTx(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	if transaction, ok := s.mempool.Get(hash); ok {
		writeJSON(w, http.StatusOK, transaction)
		return
	}
	if transaction, ok := s.indexer.FindTransaction(strings.ToLower(hash)); ok {
		writeJSON(w, http.StatusOK, transaction)
		return
	}
	writeError(w, http.StatusNotFound, "transaction not found")
}

func (s *Server) handleMempool(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.mempool.All())
}

func (s *Server) handleP2PStatus(w http.ResponseWriter, r *http.Request) {
	head := s.chain.Head()
	writeJSON(w, http.StatusOK, p2p.Status{NodeID: s.p2p.ID, Height: head.Height, Head: head.Hash})
}

func (s *Server) handleP2PBlock(w http.ResponseWriter, r *http.Request) {
	s.handleBlock(w, r)
}

func (s *Server) handleP2PTx(w http.ResponseWriter, r *http.Request) {
	var message p2p.TransactionMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, "invalid p2p transaction")
		return
	}
	if err := s.mempool.Add(message.Transaction); err != nil && !strings.Contains(err.Error(), "duplicate") {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	logEvent("p2p.tx_received", map[string]any{"from_node": message.FromNode, "hash": message.Transaction.Hash})
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (s *Server) handleP2PBlockMessage(w http.ResponseWriter, r *http.Request) {
	var message p2p.BlockMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, "invalid p2p block")
		return
	}
	if s.chain.HasBlock(message.Block.Hash) {
		writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
		return
	}
	if err := s.chain.AddBlock(message.Block); err != nil {
		s.metrics.IncRejected()
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.metrics.IncBlockImported()
	s.mempool.RemoveConfirmed(message.Block.Transactions)
	_ = s.save()
	s.reindex("p2p_block_import")
	logEvent("p2p.block_imported", map[string]any{"from_node": message.FromNode, "height": message.Block.Height, "hash": message.Block.Hash})
	s.p2p.BroadcastBlock(message.Block)
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
