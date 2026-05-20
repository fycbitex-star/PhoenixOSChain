package chain

type Genesis struct {
	Name           string            `json:"name"`
	ChainID        uint64            `json:"chainId"`
	NetworkID      uint64            `json:"networkId"`
	NativeCurrency NativeCurrency    `json:"nativeCurrency"`
	Consensus      GenesisConsensus  `json:"consensus"`
	Alloc          map[string]uint64 `json:"alloc"`
	Params         ChainParams       `json:"params"`
}

type NativeCurrency struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Decimals  uint8  `json:"decimals"`
	MaxSupply uint64 `json:"maxSupply"`
}

type GenesisConsensus struct {
	Type             string   `json:"type"`
	BlockTimeSeconds uint64   `json:"blockTimeSeconds"`
	Validators       []string `json:"validators"`
}

type ChainParams struct {
	GasLimit   uint64 `json:"gasLimit"`
	BaseFeeWei string `json:"baseFeeWei"`
}

func DefaultGenesis() Genesis {
	return Genesis{
		Name:      "PhoenixChain Devnet",
		ChainID:   20260518,
		NetworkID: 20260518,
		NativeCurrency: NativeCurrency{
			Name:      "Phoenix",
			Symbol:    "PHX",
			Decimals:  18,
			MaxSupply: 30000000,
		},
		Consensus: GenesisConsensus{
			Type:             "poa",
			BlockTimeSeconds: 3,
			Validators:       []string{},
		},
		Alloc: map[string]uint64{},
		Params: ChainParams{
			GasLimit:   30000000,
			BaseFeeWei: "1000000000",
		},
	}
}
