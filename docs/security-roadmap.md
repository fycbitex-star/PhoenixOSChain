# PhoenixChain Security Roadmap

PhoenixChain security must mature before any public value is placed on the
network.

## Private Key Safety

Ready target:

- no hardcoded private keys
- local wallet generation
- encrypted keystore
- validator key separation

Placeholder:

- hardware wallet support
- remote signer support

Acceptance criteria:

- private keys are never committed to the repository
- key files are excluded from source control

## Transaction Validation

Ready target:

- signature validation
- sender address validation
- balance validation
- nonce validation
- duplicate transaction rejection
- replay protection through chain ID

Placeholder:

- fee market hardening
- mempool eviction policy

Acceptance criteria:

- invalid transactions cannot enter blocks

## Block Validation

Ready target:

- block hash check
- previous hash check
- height check
- transaction root check
- validator signature verification
- validator authorization check

Placeholder:

- BFT commit verification
- advanced fork choice

Acceptance criteria:

- invalid or unauthorized blocks are rejected

## P2P Abuse Protection

Ready target:

- message size limits
- peer connection limits
- duplicate message handling
- malformed message rejection

Placeholder:

- peer scoring
- ban lists
- eclipse attack mitigation

Acceptance criteria:

- malformed peers cannot crash a node in standard tests

## RPC Protection

Ready target:

- request size limits
- rate limiting
- timeout handling
- error hygiene
- public/private API separation

Placeholder:

- API keys
- paid RPC tiers

Acceptance criteria:

- RPC endpoints resist basic abuse and expose no private key material

## Faucet Abuse Protection

Ready target:

- address limits
- IP limits
- cooldowns
- request logs

Placeholder:

- captcha
- identity or social verification

Acceptance criteria:

- repeated faucet abuse is throttled and observable

## Audit Requirements

Ready target:

- external audit before mainnet candidate
- internal review before public testnet
- dependency review
- cryptography review
- consensus review

Placeholder:

- formal verification

Acceptance criteria:

- audit findings are tracked, fixed, retested, and published where appropriate

## What Comes Next

Security work should continue in every version. The next coding step is to add
v0.5 storage and networking with abuse limits from the beginning.
