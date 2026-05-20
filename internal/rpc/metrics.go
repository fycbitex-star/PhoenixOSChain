package rpc

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	mu                sync.RWMutex
	RequestsTotal     uint64
	RejectedTotal     uint64
	TxsAcceptedTotal  uint64
	BlocksProduced    uint64
	BlocksImported    uint64
	FaucetRequests    uint64
	SeedRegistrations uint64
	LastBlockUnix     int64
	RPCByPath         map[string]uint64
}

func NewMetrics() *Metrics {
	return &Metrics{RPCByPath: make(map[string]uint64)}
}

func (m *Metrics) IncRequest(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RequestsTotal++
	m.RPCByPath[path]++
}

func (m *Metrics) IncRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RejectedTotal++
}

func (m *Metrics) IncTxAccepted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TxsAcceptedTotal++
}

func (m *Metrics) IncBlockProduced() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlocksProduced++
	m.LastBlockUnix = time.Now().Unix()
}

func (m *Metrics) IncBlockImported() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlocksImported++
	m.LastBlockUnix = time.Now().Unix()
}

func (m *Metrics) IncFaucetRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FaucetRequests++
}

func (m *Metrics) IncSeedRegistration() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SeedRegistrations++
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "phoenixchain_rpc_requests_total %d\n", s.metrics.RequestsTotal)
	fmt.Fprintf(w, "phoenixchain_rpc_rejected_total %d\n", s.metrics.RejectedTotal)
	fmt.Fprintf(w, "phoenixchain_transactions_accepted_total %d\n", s.metrics.TxsAcceptedTotal)
	fmt.Fprintf(w, "phoenixchain_blocks_produced_total %d\n", s.metrics.BlocksProduced)
	fmt.Fprintf(w, "phoenixchain_blocks_imported_total %d\n", s.metrics.BlocksImported)
	fmt.Fprintf(w, "phoenixchain_faucet_requests_total %d\n", s.metrics.FaucetRequests)
	fmt.Fprintf(w, "phoenixchain_seed_registrations_total %d\n", s.metrics.SeedRegistrations)
	fmt.Fprintf(w, "phoenixchain_last_block_timestamp_seconds %d\n", s.metrics.LastBlockUnix)
	fmt.Fprintf(w, "phoenixchain_chain_height %d\n", s.chain.Height())
	fmt.Fprintf(w, "phoenixchain_mempool_size %d\n", len(s.mempool.All()))
	fmt.Fprintf(w, "phoenixchain_peer_count %d\n", len(s.p2p.SnapshotPeers()))
	fmt.Fprintf(w, "phoenixchain_healthy_peer_count %d\n", s.p2p.HealthyPeerCount())
	if s.validator != "" {
		fmt.Fprintf(w, "phoenixchain_validator_enabled 1\n")
	} else {
		fmt.Fprintf(w, "phoenixchain_validator_enabled 0\n")
	}
	for path, count := range s.metrics.RPCByPath {
		fmt.Fprintf(w, "phoenixchain_rpc_path_requests_total{path=%q} %d\n", path, count)
	}
}
