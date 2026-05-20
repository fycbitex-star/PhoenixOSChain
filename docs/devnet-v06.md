# PhoenixChain v0.6 Devnet Hardening

PhoenixChain v0.6 is a professionalized devnet with explorer UI, RPC hardening,
metrics, operational scripts, and Fycbit integration. It is not production-ready
and not mainnet-ready.

## Architecture

- 3 node local validator network
- node-1 and node-2 are PoA validators
- node-3 is public RPC/explorer node
- nginx exposes `https://chain.fycbit.com`
- BadgerDB persists chain snapshots and indexes

## Explorer

Explorer pages:

- `/`
- `/blocks`
- `/txs`
- `/block-view/{height}`
- `/tx-view/{hash}`
- `/account-view/{address}`
- `/validators`

## RPC

JSON endpoints:

- `/health`
- `/chain/head`
- `/chain/latest`
- `/chain/block/{height}`
- `/account/{address}`
- `/tx/{hash}`
- `/mempool`
- `/peers`
- `/metrics`

## Security

Controls:

- nginx rate limiting
- request size limits
- application rate limiting
- malformed payload rejection
- structured logs
- localhost-only node ports
- systemd restart policy

## Monitoring

Prometheus-style metrics are available at `/metrics`.

## Validator Operations

Validators are fixed PoA devnet validators. Staking, slashing, rewards, and
governance are not active.

## Backup Strategy

Use:

- `scripts/backup-phoenixchain.sh`
- `scripts/restore-phoenixchain.sh`

## Current Limitations

- no BFT
- no staking
- no slashing
- no smart contracts
- no exchange custody
- no public testnet faucet
- no mainnet

## Next Roadmap

Move toward public testnet preparation: faucet, external validator onboarding,
off-server monitoring, CI deployment, and security review.
