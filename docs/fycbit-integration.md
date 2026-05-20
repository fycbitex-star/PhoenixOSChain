# Fycbit Integration Plan

PhoenixChain is integrated as a devnet service at:

```text
https://chain.fycbit.com
```

## Implemented

- PhoenixChain explorer UI on `chain.fycbit.com`.
- Public RPC endpoints coexist with explorer pages.
- Chain status data is available from `/health`, `/chain/head`, `/peers`, and `/metrics`.
- The existing `fycbit.com`, `api.fycbit.com`, and `evm.fycbit.com` routing remains separate.

## Status Widget

Fycbit can consume:

- latest block: `GET https://chain.fycbit.com/chain/head`
- health: `GET https://chain.fycbit.com/health`
- peers: `GET https://chain.fycbit.com/peers`
- metrics: `GET https://chain.fycbit.com/metrics`

## Deposit Architecture Placeholder

Do not enable real custody yet.

Future deposit flow:

1. generate user deposit address
2. monitor PhoenixChain blocks
3. confirm transfer depth
4. credit internal exchange ledger
5. record audit trail

## Withdrawal Architecture Placeholder

Future withdrawal flow:

1. user requests withdrawal
2. risk engine approves
3. hot wallet signs transaction
4. broadcaster submits transaction
5. ledger records tx hash and confirmation status

## Wallet Integration Strategy

Start with read-only balance viewing. Do not store production private keys in the
exchange app until custody architecture is reviewed.

## Chain Monitoring Strategy

Fycbit admin should display:

- chain height
- latest block hash
- node health
- validator health
- mempool size
- RPC availability
