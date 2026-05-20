# PhoenixOS Integration Roadmap

PhoenixOS integration comes after PhoenixChain has a stable testnet-grade API.
It must not drive premature mainnet launch.

## PhoenixOS Wallet Module

Ready target:

- create or import wallet
- show PHX balance
- send PHX
- receive PHX
- show transaction status

Placeholder:

- custody recovery
- hardware wallet integration

Acceptance criteria:

- PhoenixOS users can manage testnet PHX safely without exposing private keys

## User Wallet Mapping

Ready target:

- map PhoenixOS user ID to wallet address
- allow multiple wallets per user
- audit wallet changes

Placeholder:

- custodial wallet service
- enterprise identity integration

Acceptance criteria:

- wallet ownership changes are traceable and reversible at the product layer

## PHX Payment System

Ready target:

- payment request creation
- transaction submission
- payment confirmation
- failed payment handling

Placeholder:

- refunds
- recurring billing automation
- compliance review

Acceptance criteria:

- PhoenixOS can detect whether a PHX payment succeeded or failed

## Transaction History

Ready target:

- list user transactions
- filter sent and received transfers
- show block confirmation status

Placeholder:

- advanced analytics
- tax reports

Acceptance criteria:

- users can audit their own PhoenixOS-linked PHX activity

## Staking Panel

Ready target:

- validator list
- staking status placeholder
- delegation design placeholder

Placeholder:

- real staking actions
- slashing warnings

Acceptance criteria:

- staking UI remains disabled until protocol staking exists

## AI Agent Payment Hooks

Ready target:

- payment intent API
- PHX spend approval
- agent activity logging
- spending limits

Placeholder:

- autonomous high-value transactions
- cross-chain payments

Acceptance criteria:

- AI agents cannot spend PHX without explicit policy limits

## Subscription Payment With PHX

Ready target:

- subscription invoice
- PHX payment confirmation
- service activation after confirmation

Placeholder:

- automatic recurring on-chain pulls
- dispute workflows

Acceptance criteria:

- PhoenixOS subscriptions can be marked paid after confirmed PHX transfer

## On-Chain Activity Audit Logs

Ready target:

- user activity event
- transaction hash reference
- timestamp
- service action reference

Placeholder:

- legal compliance exports
- enterprise retention policies

Acceptance criteria:

- PhoenixOS can correlate product activity with on-chain events

## What Comes Next

PhoenixOS integration should begin after v0.9 public testnet APIs are stable.
Before that, only mock integrations and API contracts should be developed.
