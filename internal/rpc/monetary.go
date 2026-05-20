package rpc

import (
	"fmt"

	"github.com/phoenixchain/phoenixchain/internal/chain"
)

type supplyAllocation struct {
	Label      string
	Percentage string
	Scope      string
}

func nativeAssetMetadata(genesis chain.Genesis) map[string]any {
	maxSupply := genesis.NativeCurrency.MaxSupply
	return map[string]any{
		"name":                 genesis.NativeCurrency.Name,
		"symbol":               genesis.NativeCurrency.Symbol,
		"decimals":             genesis.NativeCurrency.Decimals,
		"max_supply":           maxSupply,
		"max_supply_display":   fmt.Sprintf("%s %s", formatWholeNumber(maxSupply), genesis.NativeCurrency.Symbol),
		"max_supply_compact":   fmt.Sprintf("%dM %s", maxSupply/1000000, genesis.NativeCurrency.Symbol),
		"emission_state":       "fixed-cap planned monetary architecture",
		"circulating_supply":   nil,
		"circulating_note":     "not published in current devnet runtime",
		"treasury_state":       "allocation architecture defined, unlock schedule not published",
		"staking_state":        "staking and validator rewards are planned, not active in current PoA devnet",
		"governance_utilities": []string{"gas", "staking", "governance", "treasury voting", "validator economics", "ecosystem settlement", "marketplace settlement", "contract execution"},
	}
}

func supplyAllocations() []supplyAllocation {
	return []supplyAllocation{
		{Label: "Validator / Staking Rewards", Percentage: "22%", Scope: "Reserved for future validator-secured emissions once staking activates."},
		{Label: "Treasury", Percentage: "18%", Scope: "Protocol treasury architecture placeholder for long-horizon ecosystem operations."},
		{Label: "Ecosystem Growth", Percentage: "16%", Scope: "Ecosystem expansion, infrastructure grants and chain adoption programs."},
		{Label: "Developer Incentives", Percentage: "12%", Scope: "Builder programs, tooling and protocol-native application support."},
		{Label: "Liquidity", Percentage: "10%", Scope: "Liquidity architecture placeholder; no exchange listing or market-making claim is implied."},
		{Label: "Foundation Reserve", Percentage: "9%", Scope: "Long-duration reserve for chain operations and resilience."},
		{Label: "Community Incentives", Percentage: "8%", Scope: "Community activation, educational campaigns and governance participation."},
		{Label: "Strategic Reserve", Percentage: "5%", Scope: "Strategic infrastructure reserve subject to future governance controls."},
	}
}

func formatWholeNumber(value uint64) string {
	raw := fmt.Sprintf("%d", value)
	if len(raw) <= 3 {
		return raw
	}
	out := make([]byte, 0, len(raw)+len(raw)/3)
	prefix := len(raw) % 3
	if prefix == 0 {
		prefix = 3
	}
	out = append(out, raw[:prefix]...)
	for i := prefix; i < len(raw); i += 3 {
		out = append(out, ',')
		out = append(out, raw[i:i+3]...)
	}
	return string(out)
}
