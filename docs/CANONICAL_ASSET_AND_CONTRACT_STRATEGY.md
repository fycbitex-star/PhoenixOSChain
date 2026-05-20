# Canonical Asset and Contract Strategy

This document answers one of the most important questions PhoenixChain will face:

why is there not yet a single "official contract address" for PHX?

The short answer is:

because PhoenixChain is not supposed to build a token first and then invent infrastructure around it.

It is supposed to build infrastructure first and then publish the canonical asset model with enough legitimacy that the market can trust it.

## Core Decision

The canonical PHX asset should be:

- **native on PhoenixChain**
- **governed by chain monetary rules**
- **referenced by explorer, wallet, validator, governance, staking, treasury, and RPC metadata**

It should **not** be introduced first as:

- a random external ERC-20
- a cosmetic PHX-20 contract pretending to be the root asset
- a bridge-first wrapped token without sovereign issuance discipline

## Why There Is No Official Contract Yet

There are several reasons.

### 1. The canonical asset is not meant to be "just a contract"

PHX is positioned as:

- gas
- staking asset
- governance asset
- treasury voting asset
- validator economics asset
- protocol settlement asset

That means the root form of PHX should come from chain rules, not from an application-layer token contract alone.

### 2. The launch ceremony is not finished

An official asset launch requires:

- frozen monetary architecture
- frozen genesis allocation
- signer ceremony
- treasury multisig posture
- validator agreement
- public documentation
- security review
- operational rehearseability

If those are not done, an "official contract" can create confusion rather than legitimacy.

### 3. Governance and treasury execution are still conservative

The project already exposes governance posture and treasury safety direction.

But it intentionally does **not** fake:

- live weighted quorum settlement
- live treasury execution
- final governance-driven issuance changes

Publishing a final contract too early would outrun governance maturity.

### 4. External-chain representations should come after native truth

If PHX appears on foreign networks later, those assets should be represented as:

- wrapped PHX
- bridged PHX
- exchange settlement representations

Those are downstream representations.

They are not the root asset.

## The Correct Asset Model

PhoenixChain should operate with a layered asset model.

### Layer 1: Canonical Native PHX

This is the root asset.

Properties:

- chain-native
- referenced in genesis
- canonical in explorer and RPC
- validator/staking/governance aligned
- not dependent on a contract admin for base existence

### Layer 2: PHX-20 Ecosystem Assets

These are application and ecosystem contracts.

Examples:

- stable assets
- grant instruments
- treasury program assets
- utility assets
- marketplace settlement assets

PHX-20 is critical infrastructure, but it should not be mistaken for the root PHX asset model.

### Layer 3: External Wrapped Representations

If PhoenixChain expands to other chains, official external representations may exist.

Examples:

- wPHX on an EVM network
- bridged PHX in exchange-integrated environments

These should be labeled clearly as:

- wrapped
- bridged
- non-canonical root settlement representations

## What Should Become the Official Public "Contract"

If the public asks for a contract address, the project should separate the answer into two categories.

### A. Canonical Asset Answer

The canonical PHX asset is native on PhoenixChain.

That means the "official PHX" answer is:

- chain-native issuance
- chain metadata
- genesis-defined monetary base

not a single contract address.

### B. Contract Representation Answer

If a public contract address is later published, it should be one of these:

- the official wrapped PHX contract on an external chain
- an official treasury-controlled gateway representation
- an official bridge-minted representation

That address should never be confused with the root monetary truth of PhoenixChain.

## Why PHX-20 Should Not Be the Canonical PHX Contract

PHX-20 is the right standard for fungible contracts on PhoenixChain.

It is not automatically the right home for the root asset.

If PHX itself is reduced to a governance-ready token template too early, the system weakens:

- the distinction between protocol asset and application asset collapses
- validator economics become less conceptually clean
- monetary legitimacy depends on contract ceremony instead of chain ceremony
- treasury and governance semantics become muddied

The better architecture is:

- PHX as native root asset
- PHX-20 as contract asset layer

## What Must Be Completed Before Official Asset Publication

The following should be complete before a final PHX contract or wrapped representation is announced publicly as official.

### Governance

- published governance process
- published treasury controls
- published emergency posture
- final authority for official contract publication

### Treasury

- multisig participant set
- threshold rules
- signer ceremony
- recipient policy
- auditability of official treasury actions

### Validator and Mainnet

- validator agreement
- reproducible genesis
- chain parameter freeze
- operational runbooks
- disaster recovery rehearsal

### Security

- contract review
- bridge review if external representation exists
- deploy ceremony review
- signer key handling discipline

## Recommended Launch Sequence

The right order is:

1. freeze the PHX monetary architecture
2. freeze genesis allocation inputs
3. finalize validator and treasury signer posture
4. publish canonical mainnet asset position
5. launch native PHX as chain root asset
6. only then publish official wrapped or bridged PHX representations if needed

## EVM vs Native vs Wrapped

The recommended stance is clear.

### Native PHX

This should be canonical.

### PHX-20 PHX

This may exist only if there is a very deliberate reason, and it should not replace the native identity of PHX.

### Wrapped PHX on External Chains

This is the most likely form of official public contract address that markets will reference later.

But it should be announced as:

- official wrapped representation
- not the sovereign root issuance layer

## The Strategic Message

PhoenixChain should be able to say the following without ambiguity:

PHX is not waiting for a contract because PHX is being designed as a sovereign native asset first.

If and when official contract addresses are published, they will be downstream institutional representations of a native monetary base, not substitutes for it.

That is the difference between a token-first project and an infrastructure-first chain.
