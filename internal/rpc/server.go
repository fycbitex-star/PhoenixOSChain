package rpc

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/consensus"
	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/governance"
	"github.com/phoenixchain/phoenixchain/internal/indexer"
	"github.com/phoenixchain/phoenixchain/internal/p2p"
	"github.com/phoenixchain/phoenixchain/internal/phx20"
	"github.com/phoenixchain/phoenixchain/internal/storage"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

type Server struct {
	addr         string
	chain        *chain.Blockchain
	mempool      *tx.Mempool
	p2p          *p2p.Node
	db           storage.DB
	genesis      chain.Genesis
	validatorKey ed25519.PrivateKey
	validator    string
	blockTime    time.Duration
	metrics      *Metrics
	rateLimiter  *rateLimiter
	faucetKey    ed25519.PrivateKey
	faucetAddr   string
	faucetMu     sync.Mutex
	faucetLast   map[string]time.Time
	faucetHistory []faucetGrantRecord
	publicPeers  map[string]p2p.PeerStatus
	publicPeerMu sync.RWMutex
	indexer      *indexer.RuntimeIndexer
	phx20        *phx20.Manager
	governance   *governance.Manager
}

type Options struct {
	Addr         string
	Chain        *chain.Blockchain
	Mempool      *tx.Mempool
	P2P          *p2p.Node
	DB           storage.DB
	Genesis      chain.Genesis
	ValidatorKey ed25519.PrivateKey
	BlockTime    time.Duration
	FaucetKey    ed25519.PrivateKey
	PHX20        phx20.Snapshot
	Governance   governance.Snapshot
}

func NewServer(opts Options) *Server {
	validator := ""
	if opts.ValidatorKey != nil {
		validator = phxcrypto.AddressFromPublicKey(opts.ValidatorKey.Public().(ed25519.PublicKey))
		opts.Chain.AddValidator(validator)
	}
	if opts.BlockTime == 0 {
		opts.BlockTime = 3 * time.Second
	}
	faucetAddr := ""
	if opts.FaucetKey != nil {
		faucetAddr = phxcrypto.AddressFromPublicKey(opts.FaucetKey.Public().(ed25519.PublicKey))
	}
	manager := phx20.NewManager()
	_ = manager.LoadSnapshot(opts.PHX20)
	governanceManager := governance.NewManager()
	_ = governanceManager.LoadSnapshot(opts.Governance)
	return &Server{
		addr:         opts.Addr,
		chain:        opts.Chain,
		mempool:      opts.Mempool,
		p2p:          opts.P2P,
		db:           opts.DB,
		genesis:      opts.Genesis,
		validatorKey: opts.ValidatorKey,
		validator:    validator,
		blockTime:    opts.BlockTime,
		metrics:      NewMetrics(),
		rateLimiter:  newRateLimiter(120, time.Minute),
		faucetKey:    opts.FaucetKey,
		faucetAddr:   faucetAddr,
		faucetLast:   make(map[string]time.Time),
		faucetHistory: make([]faucetGrantRecord, 0, 20),
		publicPeers:  make(map[string]p2p.PeerStatus),
		indexer:      indexer.New(),
		phx20:        manager,
		governance:   governanceManager,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.reindex("startup")
	if err := s.save(); err != nil {
		logEvent("storage.save_failed", map[string]any{"error": err.Error(), "reason": "governance_bootstrap"})
	}
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	if s.validatorKey != nil {
		logEvent("validator.enabled", map[string]any{"validator": s.validator})
		go s.blockLoop(ctx)
	}
	go s.syncLoop(ctx)

	server := &http.Server{
		Addr:              s.addr,
		Handler:           s.middleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	logEvent("node.listen", map[string]any{"addr": s.addr, "node": s.p2p.ID})
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) blockLoop(ctx context.Context) {
	ticker := time.NewTicker(s.blockTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			transactions := s.mempool.All()
			if len(transactions) == 0 {
				continue
			}
			nextHeight := s.chain.Head().Height + 1
			if !s.chain.IsProposer(s.validator, nextHeight) {
				continue
			}
			engine := consensus.NewPoA(consensus.Config{
				ValidatorAddress: s.validator,
				ValidatorKey:     s.validatorKey,
				Validators:       s.chain.Validators(),
				ValidatorList:    s.chain.ValidatorList(),
			})
			block, err := engine.BuildAndSign(s.chain.Head(), transactions)
			if err != nil {
				logEvent("block.build_failed", map[string]any{"error": err.Error()})
				continue
			}
			if err := s.chain.AddBlock(block); err != nil {
				logEvent("block.import_failed", map[string]any{"error": err.Error(), "height": block.Height})
				continue
			}
			s.metrics.IncBlockProduced()
			s.mempool.RemoveConfirmed(transactions)
			if err := s.save(); err != nil {
				logEvent("storage.save_failed", map[string]any{"error": err.Error()})
			}
			s.reindex("block_loop")
			logEvent("block.produced", map[string]any{"height": block.Height, "hash": block.Hash, "txs": len(block.Transactions), "validator": s.validator})
			s.p2p.BroadcastBlock(block)
		}
	}
}

func (s *Server) syncLoop(ctx context.Context) {
	s.syncOnce()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncOnce()
		}
	}
}

func (s *Server) syncOnce() {
	for _, peer := range s.p2p.SnapshotPeers() {
		status, err := s.p2p.FetchStatus(peer)
		if err != nil {
			continue
		}
		for height := s.chain.Height() + 1; height <= status.Height; height++ {
			block, err := s.p2p.FetchBlock(peer, height)
			if err != nil {
				break
			}
			if err := s.chain.AddBlock(block); err != nil {
				logEvent("sync.block_rejected", map[string]any{"peer": peer.Address, "height": height, "error": err.Error()})
				break
			}
			s.metrics.IncBlockImported()
			s.mempool.RemoveConfirmed(block.Transactions)
			_ = s.save()
			s.reindex("sync_import")
			logEvent("sync.block_imported", map[string]any{"peer": peer.Address, "height": height, "hash": block.Hash})
		}
	}
}

func (s *Server) save() error {
	return s.db.Save(storage.Snapshot{
		Genesis:    s.genesis,
		Blocks:     s.chain.Blocks(),
		Accounts:   s.chain.State().Snapshot(),
		PHX20:      s.phx20.Snapshot(),
		Governance: s.governance.Snapshot(),
	})
}

func (s *Server) reindex(reason string) {
	s.indexer.Rebuild(s.chain.Blocks(), s.chain.State().Snapshot(), reason)
}

func logEvent(event string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["event"] = event
	fields["time"] = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(fields)
	if err != nil {
		fmt.Fprintf(os.Stdout, `{"event":"%s","time":"%s"}`+"\n", event, time.Now().UTC().Format(time.RFC3339))
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}
