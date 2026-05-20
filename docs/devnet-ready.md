# PhoenixChain Devnet v0.5 Working Edition

PhoenixChain Devnet v0.5 is a locally working blockchain development network.
It is not production-ready and must not be used as mainnet software.

## What Works

- three local nodes can run together
- nodes connect through configured peers
- transactions can be submitted through RPC
- transactions propagate to peers
- authorized validators create signed blocks
- blocks propagate to peers
- balances and nonces update after block import
- a late-starting node catches up through basic sync
- node restart reloads chain state from BadgerDB
- CLI wallet, transaction, account, chain, and peers commands work
- Docker Compose files are present for containerized local devnet

## Architecture

The node binary is `phoenixd`.

Core packages:

- `internal/chain`: block format, genesis, validation, blockchain
- `internal/tx`: transaction model, signing, mempool
- `internal/state`: account balances and nonces
- `internal/consensus`: devnet-grade PoA prototype
- `internal/p2p`: HTTP-based peer broadcast and sync
- `internal/rpc`: JSON HTTP API
- `internal/storage`: BadgerDB snapshot persistence
- `internal/config`: devnet config loader

## How To Bootstrap

Generate local validator keys, a funded dev wallet, and a devnet genesis:

```powershell
go run ./cmd/phoenixd devnet bootstrap --genesis configs/devnet.genesis.json --env .env.devnet --wallet-dir .phoenixchain/wallets
```

The generated `.env.devnet` contains local development private keys. It is
ignored by git and must not be committed.

## How To Run Three Local Nodes

Load the generated environment variables in three PowerShell terminals, then run:

```powershell
go run ./cmd/phoenixd start --config configs/devnet-node-1.yaml
go run ./cmd/phoenixd start --config configs/devnet-node-2.yaml
go run ./cmd/phoenixd start --config configs/devnet-node-3.yaml
```

Node roles:

- `node-1`: validator
- `node-2`: validator
- `node-3`: full/RPC node

## Docker Compose

After bootstrap:

```powershell
docker compose up -d --build
```

RPC ports:

- node-1: `http://127.0.0.1:8545`
- node-2: `http://127.0.0.1:8546`
- node-3: `http://127.0.0.1:8547`

Docker must be installed locally. The repository includes `compose.yaml`,
`docker-compose.devnet.yml`, and `Dockerfile`.

## CLI Examples

List generated wallets:

```powershell
go run ./cmd/phoenixd wallet list --wallet-dir .phoenixchain/wallets
```

Create a wallet:

```powershell
go run ./cmd/phoenixd wallet new
```

Send PHX:

```powershell
go run ./cmd/phoenixd tx send --rpc http://127.0.0.1:8547 --key $env:PHX_DEV_WALLET_KEY --to ADDRESS --amount 10
```

Query chain head:

```powershell
go run ./cmd/phoenixd chain head --rpc http://127.0.0.1:8547
```

Query account:

```powershell
go run ./cmd/phoenixd account --rpc http://127.0.0.1:8547 ADDRESS
```

Query peers:

```powershell
go run ./cmd/phoenixd peers --rpc http://127.0.0.1:8547
```

## RPC API

- `GET /health`
- `GET /chain/head`
- `GET /chain/latest`
- `GET /chain/block/{height}`
- `GET /account/{address}`
- `GET /tx/{hash}`
- `GET /mempool`
- `GET /peers`
- `POST /tx/send`

Internal P2P endpoints:

- `GET /p2p/status`
- `GET /p2p/block/{height}`
- `POST /p2p/tx`
- `POST /p2p/block`

## Consensus

The devnet uses fixed-validator PoA:

- validators are listed in genesis/config
- blocks are signed by validator keys
- nodes verify validator address and signature
- round-robin proposer validation prevents both validators from producing the
  same height
- duplicate blocks are ignored
- invalid validators are rejected

This is not BFT consensus and has no slashing.

## Sync

Each node periodically checks peer status:

1. compare local height with peer height
2. fetch missing blocks by height
3. validate blocks sequentially
4. replay transactions into local state
5. persist updated snapshot to BadgerDB

If node-3 starts later, it catches up from node-1 or node-2.

## Persistence

BadgerDB stores a node snapshot containing:

- genesis
- blocks
- accounts
- balances
- nonces

Restarting a node preserves chain head and account state.

## Automated Test

Run:

```powershell
go test ./internal/devnet -run TestThreeNodeDevnetEndToEnd -v
```

The test starts three local HTTP nodes, sends a transaction, waits for block
production, verifies propagation, verifies balances, restarts node-3, and checks
that state reloads from BadgerDB.

## Current Limitations

- P2P is simple HTTP, not production gossip
- sync is sequential block fetch, not fast sync or state sync
- BadgerDB stores snapshots, not optimized block/state indexes
- no staking
- no slashing
- no smart contract VM
- no public testnet hardening
- no audit
- no mainnet

## Future Roadmap

Next phase after Devnet v0.5:

- stronger peer management
- indexed storage
- mempool reconciliation
- public testnet faucet
- explorer UI
- monitoring and metrics
- validator onboarding docs
- security hardening before public testnet
