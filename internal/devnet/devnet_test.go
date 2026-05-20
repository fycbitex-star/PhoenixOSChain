package devnet_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/p2p"
	"github.com/phoenixchain/phoenixchain/internal/rpc"
	"github.com/phoenixchain/phoenixchain/internal/storage"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func TestThreeNodeDevnetEndToEnd(t *testing.T) {
	userPub, userPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	validator1Pub, validator1Priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	validator2Pub, validator2Priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}

	user := phxcrypto.AddressFromPublicKey(userPub)
	recipient := "phxrecipient000000000000000000000000000000"
	validator1 := phxcrypto.AddressFromPublicKey(validator1Pub)
	validator2 := phxcrypto.AddressFromPublicKey(validator2Pub)

	genesis := chain.DefaultGenesis()
	genesis.Alloc[user] = 100
	genesis.Consensus.Validators = []string{validator1, validator2}
	genesis.Consensus.BlockTimeSeconds = 1

	addr1 := freeAddr(t)
	addr2 := freeAddr(t)
	addr3 := freeAddr(t)
	url1 := "http://" + addr1
	url2 := "http://" + addr2
	url3 := "http://" + addr3

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTestNode(t, ctx, "node-1", addr1, []string{url2, url3}, genesis, validator1Priv, filepath.Join(t.TempDir(), "node1"))
	startTestNode(t, ctx, "node-2", addr2, []string{url1, url3}, genesis, validator2Priv, filepath.Join(t.TempDir(), "node2"))
	waitHealthy(t, url1)
	waitHealthy(t, url2)
	assertExplorer(t, url1)

	transaction := tx.New(user, recipient, 40, 0, 21000, 1)
	if err := tx.SignTransaction(&transaction, userPriv); err != nil {
		t.Fatal(err)
	}
	postJSON(t, url1+"/tx/send", transaction)

	waitForHeight(t, url1, 1)
	waitForHeight(t, url2, 1)
	assertBalance(t, url1, user, 60)
	assertBalance(t, url2, recipient, 40)
	assertMalformedRejected(t, url1)

	node3Dir := filepath.Join(t.TempDir(), "node3")
	node3Cancel := startTestNode(t, ctx, "node-3", addr3, []string{url1, url2}, genesis, nil, node3Dir)
	waitHealthy(t, url3)
	waitForHeight(t, url3, 1)
	assertSameHead(t, url1, url2, url3)
	assertBalance(t, url3, recipient, 40)

	node3Cancel()
	time.Sleep(200 * time.Millisecond)
	node3RestartCtx, node3RestartCancel := context.WithCancel(ctx)
	defer node3RestartCancel()
	go func() {
		server := newTestServer(t, "node-3-restart", addr3, []string{url1, url2}, genesis, nil, node3Dir)
		if err := server.Start(node3RestartCtx); err != nil {
			t.Errorf("node3 restart failed: %v", err)
		}
	}()
	waitHealthy(t, url3)
	waitForHeight(t, url3, 1)
	assertBalance(t, url3, recipient, 40)
}

func TestSeedRegistryAndFaucet(t *testing.T) {
	faucetPub, faucetPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	validatorPub, validatorPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	recipientPub, _, err := phxcrypto.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	faucetAddress := phxcrypto.AddressFromPublicKey(faucetPub)
	recipient := phxcrypto.AddressFromPublicKey(recipientPub)
	validator := phxcrypto.AddressFromPublicKey(validatorPub)
	genesis := chain.DefaultGenesis()
	genesis.Alloc[faucetAddress] = 100000
	genesis.Consensus.Validators = []string{validator}

	addr := freeAddr(t)
	url := "http://" + addr
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		db := storage.NewBadgerDB(filepath.Join(t.TempDir(), "seed-faucet"))
		bc := chain.NewBlockchain(genesis)
		pool := tx.NewMempool(1000, bc.ValidateTransaction)
		server := rpc.NewServer(rpc.Options{
			Addr:         addr,
			Chain:        bc,
			Mempool:      pool,
			P2P:          p2p.NewNode("seed-1", nil),
			DB:           db,
			Genesis:      genesis,
			ValidatorKey: validatorPriv,
			FaucetKey:    faucetPriv,
			BlockTime:    200 * time.Millisecond,
		})
		if err := server.Start(ctx); err != nil {
			t.Errorf("server failed: %v", err)
		}
	}()
	waitHealthy(t, url)

	postJSON(t, url+"/seed/register", map[string]string{"node_id": "external-1", "address": "http://203.0.113.10:18547"})
	var peers []map[string]any
	getJSON(t, url+"/seed/peers", &peers)
	if len(peers) == 0 {
		t.Fatal("expected registered public peer")
	}

	postJSON(t, url+"/faucet/request", map[string]string{"address": recipient})
	waitForHeight(t, url, 1)
	assertBalance(t, url, recipient, 1000)

	resp, err := http.Post(url+"/faucet/request", "application/json", bytes.NewBufferString(`{"address":"`+recipient+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second faucet request status = %s, want 429", resp.Status)
	}
}

func assertExplorer(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("explorer status = %s", resp.Status)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("explorer content type = %s", resp.Header.Get("Content-Type"))
	}
}

func assertMalformedRejected(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Post(baseURL+"/tx/send", "application/json", bytes.NewBufferString("{bad json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed tx status = %s", resp.Status)
	}
}

func startTestNode(t *testing.T, ctx context.Context, id, addr string, peers []string, genesis chain.Genesis, validatorKey ed25519.PrivateKey, dataDir string) context.CancelFunc {
	t.Helper()
	nodeCtx, cancel := context.WithCancel(ctx)
	go func() {
		server := newTestServer(t, id, addr, peers, genesis, validatorKey, dataDir)
		if err := server.Start(nodeCtx); err != nil {
			t.Errorf("%s failed: %v", id, err)
		}
	}()
	return cancel
}

func newTestServer(t *testing.T, id, addr string, peers []string, genesis chain.Genesis, validatorKey ed25519.PrivateKey, dataDir string) *rpc.Server {
	t.Helper()
	db := storage.NewBadgerDB(dataDir)
	snapshot, err := db.Load()
	var bc *chain.Blockchain
	if err == nil {
		bc = chain.FromBlocks(snapshot.Genesis, snapshot.Blocks, snapshot.Accounts)
	} else {
		bc = chain.NewBlockchain(genesis)
	}
	pool := tx.NewMempool(1000, bc.ValidateTransaction)
	return rpc.NewServer(rpc.Options{
		Addr:         addr,
		Chain:        bc,
		Mempool:      pool,
		P2P:          p2p.NewNode(id, peers),
		DB:           db,
		Genesis:      genesis,
		ValidatorKey: validatorKey,
		BlockTime:    200 * time.Millisecond,
	})
}

func freeAddr(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func waitHealthy(t *testing.T, baseURL string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s did not become healthy", baseURL)
}

func waitForHeight(t *testing.T, baseURL string, height uint64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		head := fetchHead(t, baseURL)
		if head.Height >= height {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("%s did not reach height %d", baseURL, height)
}

func fetchHead(t *testing.T, baseURL string) chain.Block {
	t.Helper()
	var block chain.Block
	getJSON(t, baseURL+"/chain/head", &block)
	return block
}

func assertSameHead(t *testing.T, urls ...string) {
	t.Helper()
	first := fetchHead(t, urls[0])
	for _, url := range urls[1:] {
		head := fetchHead(t, url)
		if head.Height != first.Height || head.Hash != first.Hash {
			t.Fatalf("head mismatch: %s has %d/%s, want %d/%s", url, head.Height, head.Hash, first.Height, first.Hash)
		}
	}
}

func assertBalance(t *testing.T, baseURL, address string, want uint64) {
	t.Helper()
	var account struct {
		Address string `json:"address"`
		Balance uint64 `json:"balance"`
		Nonce   uint64 `json:"nonce"`
	}
	getJSON(t, baseURL+"/account/"+address, &account)
	if account.Balance != want {
		t.Fatalf("%s balance on %s = %d, want %d", address, baseURL, account.Balance, want)
	}
}

func postJSON(t *testing.T, url string, value any) {
	t.Helper()
	data, _ := json.Marshal(value)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("post %s failed: %s", url, resp.Status)
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("get %s failed: %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}

func Example_runCommands() {
	fmt.Println("go run ./cmd/phoenixd wallet new")
	fmt.Println("go run ./cmd/phoenixd start --config configs/devnet-node-1.yaml")
	// Output:
	// go run ./cmd/phoenixd wallet new
	// go run ./cmd/phoenixd start --config configs/devnet-node-1.yaml
}
