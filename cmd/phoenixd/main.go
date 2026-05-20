package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/config"
	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
	"github.com/phoenixchain/phoenixchain/internal/p2p"
	"github.com/phoenixchain/phoenixchain/internal/rpc"
	"github.com/phoenixchain/phoenixchain/internal/storage"
	"github.com/phoenixchain/phoenixchain/internal/tx"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = cmdInit(os.Args[2:])
	case "start":
		err = cmdStart(os.Args[2:])
	case "wallet":
		err = cmdWallet(os.Args[2:])
	case "tx":
		err = cmdTx(os.Args[2:])
	case "chain":
		err = cmdChain(os.Args[2:])
	case "account":
		err = cmdAccount(os.Args[2:])
	case "peers":
		err = cmdPeers(os.Args[2:])
	case "devnet":
		err = cmdDevnet(os.Args[2:])
	default:
		usage()
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "phoenixd: %v\n", err)
		os.Exit(1)
	}
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	configPath := fs.String("config", "configs/devnet-node-1.yaml", "node config path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	genesis, err := config.LoadGenesis(cfg.GenesisPath)
	if err != nil {
		return err
	}
	db := storage.NewBadgerDB(cfg.DataDir)
	bc := chain.NewBlockchain(genesis)
	for _, validator := range cfg.Validators {
		bc.AddValidator(validator)
	}
	if err := db.Save(storage.Snapshot{
		Genesis:  genesis,
		Blocks:   bc.Blocks(),
		Accounts: bc.State().Snapshot(),
	}); err != nil {
		return err
	}
	fmt.Printf("initialized %s at %s\n", cfg.NodeID, cfg.DataDir)
	return nil
}

func cmdStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	configPath := fs.String("config", "configs/devnet-node-1.yaml", "node config path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	genesis, err := config.LoadGenesis(cfg.GenesisPath)
	if err != nil {
		return err
	}
	db := storage.NewBadgerDB(cfg.DataDir)
	snapshot, err := db.Load()
	var bc *chain.Blockchain
	if err != nil {
		bc = chain.NewBlockchain(genesis)
	} else {
		bc = chain.FromBlocks(snapshot.Genesis, snapshot.Blocks, snapshot.Accounts)
	}
	for _, validator := range cfg.Validators {
		bc.AddValidator(validator)
	}

	var validatorKey ed25519.PrivateKey
	if cfg.ValidatorKeyEnv != "" {
		value := os.Getenv(cfg.ValidatorKeyEnv)
		if value != "" {
			validatorKey, err = phxcrypto.DecodePrivateKey(value)
			if err != nil {
				return err
			}
			if len(cfg.Validators) == 0 {
				bc.AddValidator(phxcrypto.AddressFromPublicKey(validatorKey.Public().(ed25519.PublicKey)))
			}
		}
	}
	var faucetKey ed25519.PrivateKey
	if value := os.Getenv("PHX_FAUCET_KEY"); value != "" {
		faucetKey, err = phxcrypto.DecodePrivateKey(value)
		if err != nil {
			return err
		}
	} else if value := os.Getenv("PHX_DEV_WALLET_KEY"); value != "" {
		faucetKey, err = phxcrypto.DecodePrivateKey(value)
		if err != nil {
			return err
		}
	}

	mempool := tx.NewMempool(5000, bc.ValidateTransaction)
	node := p2p.NewNode(cfg.NodeID, cfg.Peers)
	server := rpc.NewServer(rpc.Options{
		Addr:         cfg.RPCAddr,
		Chain:        bc,
		Mempool:      mempool,
		P2P:          node,
		DB:           db,
		Genesis:      genesis,
		ValidatorKey: validatorKey,
		FaucetKey:    faucetKey,
		BlockTime:    time.Duration(cfg.BlockTimeSeconds) * time.Second,
		PHX20:        snapshot.PHX20,
		Governance:   snapshot.Governance,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return server.Start(ctx)
}

func cmdWallet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("wallet subcommand required")
	}
	switch args[0] {
	case "new":
		fs := flag.NewFlagSet("wallet new", flag.ExitOnError)
		walletDir := fs.String("wallet-dir", ".phoenixchain/wallets", "wallet directory")
		saveWallet := fs.Bool("save", true, "save wallet metadata locally")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		pub, priv, err := phxcrypto.GenerateKeypair()
		if err != nil {
			return err
		}
		address := phxcrypto.AddressFromPublicKey(pub)
		if *saveWallet {
			if err := saveWalletFile(*walletDir, walletRecord{
				Address:    address,
				PublicKey:  phxcrypto.EncodePublicKey(pub),
				PrivateKey: phxcrypto.EncodePrivateKey(priv),
			}); err != nil {
				return err
			}
		}
		fmt.Printf("address: %s\n", address)
		fmt.Printf("public_key: %s\n", phxcrypto.EncodePublicKey(pub))
		fmt.Printf("private_key: %s\n", phxcrypto.EncodePrivateKey(priv))
		fmt.Println("warning: local development key; do not commit or reuse for production")
		return nil
	case "list":
		fs := flag.NewFlagSet("wallet list", flag.ExitOnError)
		walletDir := fs.String("wallet-dir", ".phoenixchain/wallets", "wallet directory")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return listWallets(*walletDir)
	case "address":
		fs := flag.NewFlagSet("wallet address", flag.ExitOnError)
		key := fs.String("key", "", "private key hex")
		file := fs.String("file", "", "wallet file")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *file != "" {
			record, err := loadWalletFile(*file)
			if err != nil {
				return err
			}
			fmt.Println(record.Address)
			return nil
		}
		priv, err := phxcrypto.DecodePrivateKey(*key)
		if err != nil {
			return err
		}
		fmt.Println(phxcrypto.AddressFromPublicKey(priv.Public().(ed25519.PublicKey)))
		return nil
	default:
		return fmt.Errorf("unknown wallet command %q", args[0])
	}
}

func cmdTx(args []string) error {
	if len(args) < 1 || args[0] != "send" {
		return fmt.Errorf("supported tx command: send")
	}
	fs := flag.NewFlagSet("tx send", flag.ExitOnError)
	rpcAddr := fs.String("rpc", "http://127.0.0.1:8545", "rpc address")
	key := fs.String("key", "", "sender private key hex")
	to := fs.String("to", "", "recipient address")
	amount := fs.Uint64("amount", 0, "amount")
	gasLimit := fs.Uint64("gas-limit", 21000, "gas limit placeholder")
	gasPrice := fs.Uint64("gas-price", 1, "gas price placeholder")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	priv, err := phxcrypto.DecodePrivateKey(*key)
	if err != nil {
		return err
	}
	from := tx.AddressForPrivateKey(priv)
	account, err := fetchAccount(*rpcAddr, from)
	if err != nil {
		return err
	}
	transaction := tx.New(from, *to, *amount, account.Nonce, *gasLimit, *gasPrice)
	if err := tx.SignTransaction(&transaction, priv); err != nil {
		return err
	}
	data, _ := json.Marshal(transaction)
	resp, err := http.Post(*rpcAddr+"/tx/send", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("tx rejected: %s", resp.Status)
	}
	fmt.Println(transaction.Hash)
	return nil
}

func cmdChain(args []string) error {
	if len(args) < 1 || args[0] != "head" {
		return fmt.Errorf("supported chain command: head")
	}
	fs := flag.NewFlagSet("chain head", flag.ExitOnError)
	rpcAddr := fs.String("rpc", "http://127.0.0.1:8545", "rpc address")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	return getAndPrint(*rpcAddr + "/chain/head")
}

func cmdAccount(args []string) error {
	fs := flag.NewFlagSet("account", flag.ExitOnError)
	rpcAddr := fs.String("rpc", "http://127.0.0.1:8545", "rpc address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("account address required")
	}
	return getAndPrint(*rpcAddr + "/account/" + fs.Arg(0))
}

func cmdPeers(args []string) error {
	fs := flag.NewFlagSet("peers", flag.ExitOnError)
	rpcAddr := fs.String("rpc", "http://127.0.0.1:8545", "rpc address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return getAndPrint(*rpcAddr + "/peers")
}

func cmdDevnet(args []string) error {
	if len(args) < 1 || args[0] != "bootstrap" {
		return fmt.Errorf("supported devnet command: bootstrap")
	}
	fs := flag.NewFlagSet("devnet bootstrap", flag.ExitOnError)
	genesisPath := fs.String("genesis", "configs/devnet.genesis.json", "genesis path to update")
	envPath := fs.String("env", ".env.devnet", "env file output")
	walletDir := fs.String("wallet-dir", ".phoenixchain/wallets", "wallet directory")
	fund := fs.Uint64("fund", 1000000000, "initial funded wallet balance")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	v1Pub, v1Priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		return err
	}
	v2Pub, v2Priv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		return err
	}
	userPub, userPriv, err := phxcrypto.GenerateKeypair()
	if err != nil {
		return err
	}

	v1 := phxcrypto.AddressFromPublicKey(v1Pub)
	v2 := phxcrypto.AddressFromPublicKey(v2Pub)
	user := phxcrypto.AddressFromPublicKey(userPub)

	genesis, err := config.LoadGenesis(*genesisPath)
	if err != nil {
		genesis = chain.DefaultGenesis()
	}
	genesis.Consensus.Validators = []string{v1, v2}
	if genesis.Alloc == nil {
		genesis.Alloc = map[string]uint64{}
	}
	genesis.Alloc[user] = *fund
	if err := writeJSONFile(*genesisPath, genesis, 0o644); err != nil {
		return err
	}

	records := []walletRecord{
		{Address: v1, PublicKey: phxcrypto.EncodePublicKey(v1Pub), PrivateKey: phxcrypto.EncodePrivateKey(v1Priv)},
		{Address: v2, PublicKey: phxcrypto.EncodePublicKey(v2Pub), PrivateKey: phxcrypto.EncodePrivateKey(v2Priv)},
		{Address: user, PublicKey: phxcrypto.EncodePublicKey(userPub), PrivateKey: phxcrypto.EncodePrivateKey(userPriv)},
	}
	for _, record := range records {
		if err := saveWalletFile(*walletDir, record); err != nil {
			return err
		}
	}

	env := fmt.Sprintf("PHX_NODE1_VALIDATOR_KEY=%s\nPHX_NODE1_VALIDATOR_ADDRESS=%s\nPHX_NODE2_VALIDATOR_KEY=%s\nPHX_NODE2_VALIDATOR_ADDRESS=%s\nPHX_DEV_WALLET_KEY=%s\nPHX_DEV_WALLET_ADDRESS=%s\n",
		phxcrypto.EncodePrivateKey(v1Priv),
		v1,
		phxcrypto.EncodePrivateKey(v2Priv),
		v2,
		phxcrypto.EncodePrivateKey(userPriv),
		user,
	)
	if err := os.WriteFile(*envPath, []byte(env), 0o600); err != nil {
		return err
	}
	fmt.Printf("devnet bootstrapped\nvalidator_1: %s\nvalidator_2: %s\ndev_wallet: %s\nfunded_balance: %d\nenv_file: %s\n", v1, v2, user, *fund, *envPath)
	fmt.Println("warning: generated keys are for local devnet only")
	return nil
}

type walletRecord struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func saveWalletFile(walletDir string, record walletRecord) error {
	if err := os.MkdirAll(walletDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(walletDir, record.Address+".json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func writeJSONFile(path string, value any, mode os.FileMode) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), mode)
}

func loadWalletFile(path string) (walletRecord, error) {
	var record walletRecord
	data, err := os.ReadFile(path)
	if err != nil {
		return record, err
	}
	return record, json.Unmarshal(data, &record)
}

func listWallets(walletDir string) error {
	entries, err := os.ReadDir(walletDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		record, err := loadWalletFile(filepath.Join(walletDir, entry.Name()))
		if err == nil {
			fmt.Println(record.Address)
		}
	}
	return nil
}

func fetchAccount(rpcAddr, address string) (struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}, error) {
	var account struct {
		Address string `json:"address"`
		Balance uint64 `json:"balance"`
		Nonce   uint64 `json:"nonce"`
	}
	resp, err := http.Get(rpcAddr + "/account/" + address)
	if err != nil {
		return account, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return account, fmt.Errorf("account query failed: %s", resp.Status)
	}
	return account, json.NewDecoder(resp.Body).Decode(&account)
}

func getAndPrint(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	var out any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
	return nil
}

func usage() {
	fmt.Println("phoenixd commands:")
	fmt.Println("  init --config configs/devnet-node-1.yaml")
	fmt.Println("  start --config configs/devnet-node-1.yaml")
	fmt.Println("  wallet new")
	fmt.Println("  wallet list")
	fmt.Println("  wallet address --key PRIVATE_KEY_HEX")
	fmt.Println("  tx send --rpc http://127.0.0.1:8545 --key PRIVATE_KEY_HEX --to ADDRESS --amount AMOUNT")
	fmt.Println("  chain head --rpc http://127.0.0.1:8545")
	fmt.Println("  account --rpc http://127.0.0.1:8545 ADDRESS")
	fmt.Println("  peers --rpc http://127.0.0.1:8545")
	fmt.Println("  devnet bootstrap --genesis configs/devnet.genesis.json --env .env.devnet")
}
