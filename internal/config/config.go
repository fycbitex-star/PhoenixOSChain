package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/chain"
)

type Config struct {
	NodeID           string
	DataDir          string
	RPCAddr          string
	P2PAddr          string
	Peers            []string
	Validators       []string
	Role             string
	ValidatorKeyEnv  string
	BlockTimeSeconds uint64
	GenesisPath      string
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	cfg := Config{BlockTimeSeconds: 3}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return Config{}, fmt.Errorf("invalid config line: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := os.ExpandEnv(strings.Trim(strings.TrimSpace(parts[1]), `"`))
		switch key {
		case "node_id":
			cfg.NodeID = value
		case "data_dir":
			cfg.DataDir = value
		case "rpc_addr":
			cfg.RPCAddr = value
		case "p2p_addr":
			cfg.P2PAddr = value
		case "role":
			cfg.Role = value
		case "validator_key_env":
			cfg.ValidatorKeyEnv = value
		case "genesis":
			cfg.GenesisPath = value
		case "block_time_seconds":
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return Config{}, err
			}
			cfg.BlockTimeSeconds = parsed
		case "peers":
			cfg.Peers = splitList(value)
		case "validators":
			cfg.Validators = splitList(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	if cfg.NodeID == "" || cfg.DataDir == "" || cfg.RPCAddr == "" || cfg.P2PAddr == "" {
		return Config{}, fmt.Errorf("node_id, data_dir, rpc_addr, and p2p_addr are required")
	}
	if cfg.GenesisPath == "" {
		cfg.GenesisPath = "configs/devnet.genesis.json"
	}
	return cfg, nil
}

func LoadGenesis(path string) (chain.Genesis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return chain.Genesis{}, err
	}
	var genesis chain.Genesis
	if err := json.Unmarshal(data, &genesis); err != nil {
		return chain.Genesis{}, err
	}
	if genesis.Name == "" || genesis.ChainID == 0 || genesis.NativeCurrency.Symbol == "" {
		return chain.Genesis{}, fmt.Errorf("invalid genesis")
	}
	if genesis.Alloc == nil {
		genesis.Alloc = map[string]uint64{}
	}
	return genesis, nil
}

func splitList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"`)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
