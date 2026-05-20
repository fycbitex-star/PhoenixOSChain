# PhoenixChain Protocol v0.1

## Network

PhoenixChain v0.1 runs as a local devnet. Each node has a node ID, listen
address, RPC address, peer list, and optional validator key.

## Blocks

Blocks must satisfy:

- height increments by one
- previous hash matches the parent block
- transaction root matches the included transactions
- block hash matches the unsigned header
- validator address is in the configured validator list
- validator signature verifies against the block hash

## Transactions

Transactions must satisfy:

- hash matches the unsigned transaction payload
- signature verifies against the included public key
- sender address matches the public key
- sender has sufficient balance
- nonce equals the sender account nonce
- duplicate transaction hashes are rejected

## State Transitions

For a valid transfer:

1. subtract amount from sender balance
2. increment sender nonce
3. add amount to recipient balance

Gas limit and gas price are placeholders in v0.1 and are not charged yet.

## Consensus

PoA v0.1 uses a fixed validator set. Round-robin proposer selection is a
placeholder and not yet enforced. No slashing, finality gadget, or BFT voting is
implemented.
