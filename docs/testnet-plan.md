# PhoenixChain v0.9 Public Testnet Plan

The public testnet validates PhoenixChain with external users, validators, and
wallets. Test PHX has no real value.

## Public RPC

Ready target:

- public HTTPS RPC endpoint
- request logging
- uptime checks
- basic rate limits

Placeholder:

- paid RPC tiers
- enterprise SLAs

Acceptance criteria:

- users can query chain data and submit test transactions
- endpoint health is visible in monitoring

## Explorer

Ready target:

- blocks
- transactions
- accounts
- validator list
- network stats

Placeholder:

- advanced analytics
- contract verification

Acceptance criteria:

- users can inspect recent blocks, transactions, and balances

## Faucet

Ready target:

- test PHX distribution
- address-based limits
- IP-based limits
- abuse logging

Placeholder:

- identity verification
- advanced anti-sybil systems

Acceptance criteria:

- users can request test PHX
- repeated abuse is rate limited

## Validator Onboarding

Ready target:

- validator hardware guide
- installation guide
- key generation guide
- config examples
- upgrade process

Placeholder:

- staking economics
- slashing enforcement

Acceptance criteria:

- external validators can join testnet using documentation

## Monitoring

Ready target:

- dashboard
- uptime checks
- peer count
- block height
- block time
- RPC latency
- validator signing activity

Placeholder:

- full security operations center
- automated governance response

Acceptance criteria:

- network operators can detect stalled blocks, validator downtime, and RPC failure

## Wallet Integration

Ready target:

- wallet network metadata
- PHX balance display
- transaction submission
- test PHX flow

Placeholder:

- custody recovery
- mobile wallet certification

Acceptance criteria:

- community users can connect a wallet and send test PHX

## Community Testing

Ready target:

- issue templates
- test tasks
- public status page
- release notes

Placeholder:

- bug bounty payouts
- mainnet incentives

Acceptance criteria:

- bugs are reproducible, triaged, and assigned

## What Comes Next

After v0.9, the project must harden security, prepare audits, finalize
tokenomics, and build a mainnet candidate checklist.
