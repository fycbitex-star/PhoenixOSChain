# PhoenixChain Validator Operations

PhoenixChain v0.6 validators are fixed PoA devnet validators.

## Config Management

Validator config lives in:

- `/opt/phoenixchain/configs/server-node-1.yaml`
- `/opt/phoenixchain/configs/server-node-2.yaml`

## Key Management

- keys are loaded from `/etc/phoenixchain/devnet.env`
- file mode should be `0600`
- keys are devnet-only
- future work: encrypted keystore or remote signer

## Monitoring

Validator status is visible through:

- explorer `/validators`
- metrics `/metrics`
- systemd status
- journal logs

## Rotation Placeholder

Validator rotation requires updating genesis/config and restarting nodes. On-chain
rotation is future work.

## Rewards And Slashing Placeholder

No rewards or slashing are active in v0.6. These require protocol design before
implementation.
