package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type Node struct {
	ID     string
	Peers  []Peer
	client *http.Client
	mu     sync.RWMutex
	stats  map[string]PeerStatus
}

type PeerStatus struct {
	Address       string `json:"address"`
	Healthy       bool   `json:"healthy"`
	Failures      uint64 `json:"failures"`
	LastError     string `json:"last_error,omitempty"`
	LastSeenUnix  int64  `json:"last_seen_unix,omitempty"`
	LastCheckUnix int64  `json:"last_check_unix,omitempty"`
}

func NewNode(id string, peers []string) *Node {
	seen := make(map[string]bool)
	nodePeers := make([]Peer, 0, len(peers))
	for _, peer := range peers {
		if peer == "" || seen[peer] {
			continue
		}
		seen[peer] = true
		nodePeers = append(nodePeers, Peer{Address: peer})
	}
	return &Node{
		ID:     id,
		Peers:  nodePeers,
		client: &http.Client{Timeout: 2 * time.Second},
		stats:  make(map[string]PeerStatus),
	}
}

func (n *Node) AddPeer(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, peer := range n.Peers {
		if peer.Address == address {
			return
		}
	}
	n.Peers = append(n.Peers, Peer{Address: address})
}

func (n *Node) BroadcastTransaction(transaction tx.Transaction) {
	fmt.Printf(`{"event":"p2p.broadcast_tx","node":"%s","hash":"%s","peers":%d}`+"\n", n.ID, transaction.Hash, len(n.SnapshotPeers()))
	n.broadcast("/p2p/tx", TransactionMessage{FromNode: n.ID, Transaction: transaction})
}

func (n *Node) BroadcastBlock(block chain.Block) {
	fmt.Printf(`{"event":"p2p.broadcast_block","node":"%s","height":%d,"hash":"%s","peers":%d}`+"\n", n.ID, block.Height, block.Hash, len(n.SnapshotPeers()))
	n.broadcast("/p2p/block", BlockMessage{FromNode: n.ID, Block: block})
}

func (n *Node) FetchStatus(peer Peer) (Status, error) {
	var status Status
	if err := n.getJSON(peer.Address+"/p2p/status", &status); err != nil {
		n.recordPeer(peer.Address, false, err)
		return Status{}, err
	}
	n.recordPeer(peer.Address, true, nil)
	return status, nil
}

func (n *Node) FetchBlock(peer Peer, height uint64) (chain.Block, error) {
	var block chain.Block
	if err := n.getJSON(fmt.Sprintf("%s/p2p/block/%d", peer.Address, height), &block); err != nil {
		return chain.Block{}, err
	}
	return block, nil
}

func (n *Node) SnapshotPeers() []Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	peers := make([]Peer, len(n.Peers))
	copy(peers, n.Peers)
	return peers
}

func (n *Node) SnapshotPeerStatuses() []PeerStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	statuses := make([]PeerStatus, 0, len(n.Peers))
	for _, peer := range n.Peers {
		status := n.stats[peer.Address]
		if status.Address == "" {
			status.Address = peer.Address
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func (n *Node) HealthyPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	count := 0
	for _, status := range n.stats {
		if status.Healthy {
			count++
		}
	}
	return count
}

func (n *Node) broadcast(path string, message any) {
	n.mu.RLock()
	peers := make([]Peer, len(n.Peers))
	copy(peers, n.Peers)
	n.mu.RUnlock()

	for _, peer := range peers {
		data, err := json.Marshal(message)
		if err != nil {
			continue
		}
		_, _ = n.client.Post(peer.Address+path, "application/json", bytes.NewReader(data))
	}
}

func (n *Node) recordPeer(address string, healthy bool, err error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	status := n.stats[address]
	status.Address = address
	status.Healthy = healthy
	status.LastCheckUnix = time.Now().Unix()
	if healthy {
		status.LastSeenUnix = status.LastCheckUnix
		status.LastError = ""
	} else {
		status.Failures++
		if err != nil {
			status.LastError = err.Error()
		}
	}
	n.stats[address] = status
}

func (n *Node) getJSON(url string, out any) error {
	resp, err := n.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("peer returned %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
