# PhoenixChain PCVM Architecture

PCVM is the PhoenixChain Virtual Machine research layer. It is functional at
dev/test level, but it is not Ethereum-compatible and not production-ready.

## Implemented

- bytecode execution engine
- opcode dispatch
- gas metering
- contract deployment
- contract calls
- contract storage
- execution receipts
- event logs
- revert handling
- basic sandboxing through gas limits and isolated staged writes

## Opcode Set

- `STOP`
- `PUSH`
- `ADD`
- `SUB`
- `STORE`
- `LOAD`
- `LOG`
- `RETURN`
- `REVERT`

## Execution Model

Contract writes are staged during execution. They are committed only after a
successful stop or return. Reverts and out-of-gas failures discard staged writes.

## Storage

The current storage model is key/value per contract address. A contract state
trie is a future upgrade.

## Limitations

- no full Ethereum compatibility
- no ABI
- no contract calls between contracts
- no precompiles
- no deterministic gas schedule review
- no audit
