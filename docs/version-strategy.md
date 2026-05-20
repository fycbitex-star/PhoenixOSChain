# PhoenixChain Version Strategy

PhoenixChain uses staged versions to avoid premature production claims.

## v0.1 Core Prototype

Purpose: prove core mechanics.

Ready:

- block structure
- transaction format
- wallet/key system
- signatures
- mempool
- state database in memory
- account balances and nonces
- basic RPC and CLI
- local single-node chain

Placeholder:

- persistent storage
- P2P networking
- production consensus
- smart contracts

Acceptance criteria:

- core tests pass
- signed transactions are validated
- blocks update state correctly

Next:

- add persistence and local multi-node devnet features

## v0.5 Devnet

Purpose: prove local network behavior.

Ready:

- three local nodes
- validator logic
- transaction broadcast
- block propagation
- persistent storage
- basic sync
- node configs
- devnet scripts
- Docker Compose setup

Placeholder:

- public network safety
- advanced peer scoring
- economic validator security

Acceptance criteria:

- nodes can restart without losing chain state
- blocks propagate across all local nodes
- transactions submitted to one node can be included by a validator

Next:

- prepare public testnet tooling

## v0.9 Public Testnet

Purpose: test PhoenixChain with external users and validators.

Ready:

- faucet
- explorer
- public RPC
- validator onboarding
- monitoring
- logs
- metrics
- network health checks
- security hardening

Placeholder:

- real-value asset security
- final governance
- final staking economics

Acceptance criteria:

- external validators can join
- users can request test PHX
- explorer and public RPC remain stable under test load
- security issues are triaged before mainnet candidate

Next:

- audit preparation and mainnet candidate work

## v1.0 Mainnet Candidate

Purpose: produce a reviewed candidate that may become mainnet.

Ready:

- audit preparation
- tokenomics finalization
- staking
- slashing
- governance
- treasury
- validator documentation
- RPC redundancy
- backup strategy
- incident response

Placeholder:

- final launch requires explicit approval after audits and rehearsals

Acceptance criteria:

- audit issues are remediated or explicitly accepted
- genesis allocation is frozen
- validator launch process is rehearsed
- incident response is tested

Next:

- mainnet launch decision or another release candidate

## v1.1 PhoenixOS Integration

Purpose: connect PhoenixChain to PhoenixOS product surfaces.

Ready:

- PhoenixOS wallet module
- user wallet mapping
- PHX payment flows
- transaction history
- staking panel
- AI agent payment hooks
- subscription payment with PHX
- on-chain activity audit logs

Placeholder:

- product-specific compliance review
- custody model
- account recovery

Acceptance criteria:

- PhoenixOS can read balances and transaction history
- users can initiate PHX payments safely
- payment and audit logs are traceable

Next:

- improve user experience, custody options, and operational controls

## v2.0 Smart Contract VM / Advanced Protocol

Purpose: add programmable execution and stronger protocol features.

Ready:

- smart contract VM design
- gas model
- contract state storage
- developer tooling
- protocol upgrade governance
- advanced consensus improvements

Placeholder:

- VM choice remains open until v1.x stability

Acceptance criteria:

- VM security model is documented
- contracts run deterministically
- contract execution has test coverage and audit review

Next:

- ecosystem tooling, SDKs, and long-term protocol upgrades
