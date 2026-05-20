# Run A PhoenixChain Testnet Validator

Validator onboarding is currently permissioned for v0.7.

## Requirements

- Linux server
- static public IP
- `phoenixd` binary
- validator key generated locally
- monitoring endpoint

## Steps

1. Generate a validator wallet.
2. Send validator address and metadata to the PhoenixChain operator.
3. Wait for inclusion in validator config.
4. Run `phoenixd start --config validator.yaml`.
5. Monitor `/metrics` and systemd logs.

## Current Limitation

Validator set updates are not on-chain yet. This is not decentralized validator
admission.
