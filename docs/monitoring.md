# PhoenixChain v0.6 Monitoring

PhoenixChain exposes Prometheus-style metrics at:

```text
https://chain.fycbit.com/metrics
```

## Metrics

- `phoenixchain_rpc_requests_total`
- `phoenixchain_rpc_rejected_total`
- `phoenixchain_transactions_accepted_total`
- `phoenixchain_blocks_produced_total`
- `phoenixchain_blocks_imported_total`
- `phoenixchain_last_block_timestamp_seconds`
- `phoenixchain_chain_height`
- `phoenixchain_mempool_size`
- `phoenixchain_peer_count`
- `phoenixchain_healthy_peer_count`
- `phoenixchain_validator_enabled`

## Logs

Nodes emit structured JSON logs through systemd journal.

Useful commands:

```bash
journalctl -u phoenixchain-node@1 -f
journalctl -u phoenixchain-node@2 -f
journalctl -u phoenixchain-node@3 -f
```

## Uptime Checks

Recommended checks:

- `GET /health`
- `GET /chain/head`
- `GET /peers`
- `GET /metrics`

## Grafana Placeholder

Grafana dashboards should track chain height, block production, RPC request
volume, rejected requests, mempool size, and peer health.
