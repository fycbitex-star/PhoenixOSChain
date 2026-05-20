package rpc

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/p2p"
)

type seedRegisterRequest struct {
	Address string `json:"address"`
	NodeID  string `json:"node_id"`
}

func (s *Server) handleSeedStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"seed":           true,
		"node":           s.p2p.ID,
		"height":         s.chain.Height(),
		"public_peers":   len(s.publicPeers),
		"internal_peers": len(s.p2p.SnapshotPeers()),
	})
}

func (s *Server) handleSeedPeers(w http.ResponseWriter, r *http.Request) {
	s.publicPeerMu.RLock()
	defer s.publicPeerMu.RUnlock()
	peers := make([]p2p.PeerStatus, 0, len(s.publicPeers)+len(s.p2p.SnapshotPeers()))
	for _, peer := range s.p2p.SnapshotPeerStatuses() {
		peers = append(peers, peer)
	}
	for _, peer := range s.publicPeers {
		peers = append(peers, peer)
	}
	writeJSON(w, http.StatusOK, peers)
}

func (s *Server) handleSeedRegister(w http.ResponseWriter, r *http.Request) {
	var req seedRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid seed registration")
		return
	}
	if req.Address == "" || req.NodeID == "" {
		writeError(w, http.StatusBadRequest, "address and node_id are required")
		return
	}
	s.publicPeerMu.Lock()
	s.publicPeers[req.Address] = p2p.PeerStatus{
		Address:       req.Address,
		Healthy:       true,
		LastSeenUnix:  time.Now().Unix(),
		LastCheckUnix: time.Now().Unix(),
	}
	s.publicPeerMu.Unlock()
	s.metrics.IncSeedRegistration()
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "registered": req.Address})
}
