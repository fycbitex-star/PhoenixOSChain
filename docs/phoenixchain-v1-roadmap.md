# PhoenixChain v1 Roadmap

PhoenixChain is moving from hardened devnet toward public deployment readiness.

## Architecture

- modular Go blockchain core
- PoA devnet
- PCVM research VM
- encrypted wallet layer
- economic core
- explorer and RPC

## Consensus

Current PoA remains devnet-only. BFT consensus is a future milestone.

## VM

PCVM supports basic bytecode execution, storage, receipts, gas, logs, and
reverts. It is not Ethereum-compatible yet.

## Explorer

Explorer v0.6 provides blocks, transactions, accounts, validators, peers, and
metrics. Contract pages are next.

## Economic Layer

Validator registration, bonding, rewards, treasury accounting, and slashing
placeholder logic exist at dev/test level.

## Wallet System

Encrypted local keystore and signing API exist. Production custody is not ready.

## PhoenixOS Integration

Start with read-only status widgets and wallet views. Payments require policy,
custody, and security design.

## Public Testnet Strategy

Before public testnet:

- faucet
- external validator onboarding
- monitoring dashboard
- abuse prevention
- stronger sync and peer management

## Audit Roadmap

No audit yet. Audit requires threat model, fuzzing, static analysis, dependency
scan, and stable protocol scope.

## Mainnet Readiness Conditions

- public testnet passes
- audit findings remediated
- tokenomics finalized
- governance documented
- incident response rehearsed
- validator set confirmed
