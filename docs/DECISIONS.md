# PhoenixChain Decisions

## D001: PhoenixChain Is An Independent L1

PhoenixChain will not be deployed under Ethereum as an L2, rollup, sidechain
settled to Ethereum, or token contract. It will run its own validator network and
produce its own canonical blocks.

## D002: Ethereum Is A Compatibility Reference

Ethereum is useful as a reference because wallets, explorers, smart contract
tooling, and JSON-RPC conventions already exist. PhoenixChain should aim to be
EVM-compatible unless a specific mainnet reason requires divergence.

## D003: Production Client Language

The production client target is Go. This aligns well with the Ethereum ecosystem,
especially go-ethereum concepts and libraries, while still allowing PhoenixChain
to own its consensus, genesis, and network rules.

## D004: Consensus Path

PhoenixChain should start with a Proof of Authority devnet because it is simple
to operate and debug. The public testnet should then move toward validator-based
Proof of Stake before mainnet.

## D005: Native Asset

The native asset symbol is `PHX` unless changed later. `PHX` is the gas token,
validator staking asset, and protocol reward unit.

## D006: Native Monetary Cap

PhoenixChain locks the canonical PHX maximum supply at `30,000,000 PHX` with
`18` decimals. This value is the official reference for explorer metadata,
wallet metadata, RPC/network metadata, governance positioning, validator
economics and treasury planning.
