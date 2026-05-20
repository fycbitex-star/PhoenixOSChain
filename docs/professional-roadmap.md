# PhoenixChain Professional Roadmap

PhoenixChain must grow in controlled phases: Prototype, Devnet, Public Testnet,
Audit Preparation, and Mainnet Candidate. This roadmap does not claim production
readiness and does not authorize a mainnet launch.

## Phase 1: v0.1 Core Prototype

Goal: prove the core blockchain mechanics in a small, testable local node.

Ready in this phase:

- block structure
- transaction model
- wallet/key generation
- digital signatures
- account balances and nonces
- mempool
- basic RPC
- basic CLI
- local single-node chain

Placeholders:

- persistent storage
- real P2P sync
- multi-validator safety
- metrics and monitoring
- staking and slashing
- smart contracts

Acceptance criteria:

- a wallet can create a signed transaction
- a node can validate and include the transaction in a block
- account balances and nonces update correctly
- invalid signatures, invalid nonces, overspending, and duplicate transactions are rejected
- tests cover block hashing, transaction signing, state transition, and mempool duplicate protection

What comes next:

- upgrade the prototype into v0.5 Devnet with persistent storage and a local 3-node network

## Phase 2: v0.5 Devnet

Goal: run PhoenixChain across multiple local nodes with validator behavior,
transaction broadcast, block broadcast, persistent storage, and basic sync.

Ready in this phase:

- local 3-node network
- validator node mode
- RPC/full node mode
- bootnode placeholder
- transaction propagation
- block propagation
- persistent local storage
- basic sync from peers
- Docker Compose devnet

Placeholders:

- adversarial peer scoring
- production-grade gossip
- BFT finality
- validator rotation
- public RPC hardening

Acceptance criteria:

- three local nodes start from configs
- a transaction submitted to one node reaches validator nodes
- a validator creates and signs a block
- other nodes verify and import the block
- restarting a node preserves chain data
- a fresh node can sync blocks from a peer

What comes next:

- prepare v0.9 Public Testnet infrastructure and external operator documentation

## Phase 3: v0.9 Public Testnet

Goal: expose a safe public network for community testing without mainnet value.

Ready in this phase:

- public RPC
- explorer
- faucet
- validator onboarding process
- public node documentation
- monitoring dashboard
- uptime checks
- wallet integration
- test PHX distribution
- community testing process

Placeholders:

- real economic security
- final tokenomics
- treasury operations
- mainnet validator guarantees

Acceptance criteria:

- external users can connect wallets to the public testnet
- faucet distributes test PHX with abuse limits
- explorer shows blocks, transactions, and accounts
- validators can join using documented steps
- network health is visible through dashboards
- incidents are tracked and resolved through a written process

What comes next:

- security hardening, audit preparation, and mainnet candidate planning

## Phase 4: Audit Preparation

Goal: freeze critical protocol behavior enough for meaningful review.

Ready in this phase:

- threat model
- protocol specification
- deterministic builds
- test coverage report
- fuzzing targets
- validator operations documentation
- RPC abuse controls
- incident response runbooks

Placeholders:

- mainnet launch date
- final genesis allocation
- production validator set

Acceptance criteria:

- auditors can build, run, and test the node
- all critical protocol paths have tests
- known risks are documented
- unresolved issues are tracked with severity and owner

What comes next:

- v1.0 Mainnet Candidate only after audit feedback and remediation

## Phase 5: v1.0 Mainnet Candidate

Goal: prepare a candidate release that may become mainnet only after review,
testing, and governance approval.

Ready in this phase:

- audit remediations
- finalized chain parameters
- finalized genesis allocation
- validator launch documentation
- backup RPC strategy
- disaster recovery plan
- DDoS protection plan
- treasury multisig process
- incident response team and runbooks

Placeholders:

- mainnet activation remains gated
- governance may still reject launch
- post-launch upgrade process must be rehearsed

Acceptance criteria:

- launch checklist is complete
- validator set is verified
- genesis file is reproducible
- emergency halt and recovery procedures are rehearsed
- RPC redundancy and backups are tested
- security sign-off is recorded

What comes next:

- controlled launch decision, followed by post-launch operations and PhoenixOS integration
