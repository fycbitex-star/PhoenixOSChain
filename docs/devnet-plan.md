# PhoenixChain v0.5 Devnet Plan

The v0.5 Devnet proves PhoenixChain can run as a local multi-node network. It is
not a public testnet and not production-ready.

## Network Topology

Ready target:

- node-1: validator node
- node-2: validator node
- node-3: RPC/full node
- bootnode: peer discovery placeholder
- archive node: placeholder only

Placeholder:

- production peer discovery
- archive indexing
- public network security

Acceptance criteria:

- all three nodes start locally from config files
- each node has a unique node identity
- nodes can connect to configured peers

## Persistent Storage

Ready target:

- block storage
- account state storage
- transaction index storage
- node restart recovery

Placeholder:

- state pruning
- snapshots
- archive mode
- database migrations

Acceptance criteria:

- node restart preserves chain head and account state
- corrupted data is detected and reported clearly

## Propagation

Ready target:

- transaction broadcast
- block broadcast
- receive transaction
- receive block
- reject invalid transaction or block

Placeholder:

- complex gossip
- peer scoring
- mempool reconciliation

Acceptance criteria:

- transaction sent to node-3 reaches validators
- validator block reaches all nodes
- duplicate transactions and blocks are ignored

## Basic Sync

Ready target:

- peer height check
- missing block request
- sequential block import

Placeholder:

- fast sync
- state sync
- fork choice beyond longest valid local chain

Acceptance criteria:

- a fresh node can catch up from an existing node
- invalid blocks are rejected during sync

## CLI Commands

Ready target:

- `phoenixd init`
- `phoenixd start --config configs/devnet-node-1.yaml`
- `phoenixd wallet new`
- `phoenixd wallet address`
- `phoenixd tx send --to ADDRESS --amount AMOUNT`
- `phoenixd chain head`
- `phoenixd account ADDRESS`

Placeholder:

- encrypted keystore
- remote signing
- hardware wallet support

Acceptance criteria:

- a developer can run the whole local devnet from documented commands

## Docker Compose Setup

Ready target:

- one compose file for three local nodes
- persistent volumes per node
- exposed RPC port for node-3
- reset script for local state

Placeholder:

- Kubernetes deployment
- production secrets management

Acceptance criteria:

- `docker compose up` starts the devnet
- `scripts/reset-devnet.sh` resets local volumes safely

## What Comes Next

After v0.5, PhoenixChain should move toward v0.9 Public Testnet with public RPC,
faucet, explorer, onboarding docs, and monitoring.
