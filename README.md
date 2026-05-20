# PhoenixOSChain

PhoenixOSChain is the public repository for **PhoenixChain**: a financial-infrastructure-first Layer 1 designed to sit beside PhoenixOS as its sovereign settlement, governance, validator, treasury, and asset runtime.

This repository is meant to do two things at once:

1. make the architecture visible enough that serious operators, builders, and researchers immediately understand the ambition
2. keep commercial control, derivative rights, and product ownership with the PhoenixOS/PhoenixChain rights holder

## What PhoenixChain Is

PhoenixChain is being built as a chain for:

- validator-secured financial infrastructure
- governance-aware treasury coordination
- PHX-native economic policy
- PHX-20 asset issuance and lifecycle support
- wallet, explorer, validator, and execution interoperability with PhoenixOS

It is **not** positioned as a meme chain, marketing-only chain, or thin wrapper around an existing dashboard product.

## Why This Exists

PhoenixOS grew beyond a single trading interface.

Once execution, OMS, wallet, governance, AI, replay, collaboration, and treasury coordination became part of the same operating system, the platform needed a chain layer that could eventually carry:

- sovereign asset logic
- validator identity
- treasury governance
- protocol decisions
- public financial infrastructure credibility

PhoenixChain is that layer.

## Design Direction

- sovereign chain, not an Ethereum L2
- Ethereum-style compatibility where it improves operator and wallet ergonomics
- native monetary identity through `PHX`
- validator-first network posture
- governance and treasury hardening as first-class system concerns
- truth-first visibility over fake decentralization or fake performance claims

## Current State

PhoenixChain is **advanced devnet / governance-runtime prototype software**.

It includes real code for:

- chain state and block flow
- validator-facing devnet operations
- governance runtime foundations
- PHX-20 runtime surfaces
- RPC and explorer-style routes
- wallet/account primitives
- security and operational documentation

It is **not yet a production mainnet**.

## Repository Structure

- `cmd/phoenixd`
  node binary and CLI entrypoint
- `internal/chain`
  block, genesis, validation, and chain state flow
- `internal/consensus`
  current consensus foundations
- `internal/governance`
  proposal, audit, and governance runtime foundations
- `internal/phx20`
  PHX-20 asset runtime
- `internal/rpc`
  RPC, explorer, governance, validator, and wallet-facing HTTP surfaces
- `internal/wallet`
  wallet and keystore primitives
- `internal/economics`
  monetary and tokenomics primitives
- `configs`
  devnet and chain configuration
- `docs`
  architecture, operations, security, validator, governance, and launch material
- `deploy`
  deploy scaffolding

## Strategic Pillars

### 1. Validator-Centered Security

PhoenixChain is intentionally designed to feel like a chain with operators, health, lifecycle, and accountability. Validator surfaces are not treated as decorative explorer add-ons.

### 2. Governance as Runtime, Not Decoration

Proposal lifecycle, voting posture, treasury safety, audit trail, and execution delay are treated as chain-level responsibilities, not just a DAO UI skin.

### 3. Monetary Clarity

The project prefers explicit monetary architecture over vague token marketing. PHX is positioned as a protocol-native asset with clear infrastructure role.

### 4. Institutional Operating Language

The repository and public documentation aim to communicate like infrastructure, not hype:

- what exists
- what is partial
- what is blocked
- what still requires audit, rehearsal, and launch discipline

## What Makes It Interesting

PhoenixChain is not only a chain repository. It is the chain half of a larger operating civilization:

- PhoenixOS handles execution, AI, replay, wallet, OMS, and operator tooling
- PhoenixChain carries the network, validator, governance, treasury, and asset trust boundary

That pairing is the real differentiator.

## Build

```powershell
go test ./...
go run ./cmd/phoenixd wallet new
go run ./cmd/phoenixd init --config configs/devnet-node-1.yaml
go run ./cmd/phoenixd start --config configs/devnet-node-1.yaml
```

## Devnet Bootstrap

```powershell
go run ./cmd/phoenixd devnet bootstrap --genesis configs/devnet.genesis.json --env .env.devnet --wallet-dir .phoenixchain/wallets
```

## Public Documentation Map

Start here:

- [docs/WHY_PHOENIXCHAIN.md](docs/WHY_PHOENIXCHAIN.md)
- [docs/ARCHITECTURE_OVERVIEW.md](docs/ARCHITECTURE_OVERVIEW.md)
- [docs/GOVERNANCE_AND_TREASURY.md](docs/GOVERNANCE_AND_TREASURY.md)
- [docs/VALIDATOR_NETWORK.md](docs/VALIDATOR_NETWORK.md)
- [docs/SECURITY_POSTURE.md](docs/SECURITY_POSTURE.md)
- [docs/MAINNET_DISCIPLINE.md](docs/MAINNET_DISCIPLINE.md)

## License

This repository is **source-available, not open source**.

- Code is licensed under [PolyForm Strict 1.0.0](LICENSE).
- Commercial rights, redistribution rights, and derivative product rights are not granted here.
- Trademark, product identity, and branding rights are reserved.

If you want commercial, OEM, derivative, or redistribution rights, obtain a separate written license from the rights holder.

## Final Note

PhoenixOSChain is published to make the direction obvious.

The goal is not to hide ambition.
The goal is to publish it clearly enough that serious people immediately understand the system, while keeping ownership and commercial control where it belongs.
