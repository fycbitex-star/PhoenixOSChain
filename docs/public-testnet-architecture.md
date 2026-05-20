# PhoenixChain v0.7 Public Testnet Architecture

PhoenixChain v0.7 is a public test network, not mainnet.

## Components

- public explorer and RPC: `https://chain.fycbit.com`
- seed registry: `/seed/status`, `/seed/peers`, `/seed/register`
- faucet: `/faucet`, `/faucet/request`
- validators: fixed PoA testnet validators
- public full nodes: external operators can register with the seed endpoint
- monitoring: `/metrics`

## Bootstrap Flow

1. operator installs `phoenixd`
2. operator downloads testnet config
3. node queries seed peers
4. node connects to peers
5. node syncs blocks
6. optional validator registration is reviewed manually

## Security Boundary

Public users can access node-3 through nginx. Validator nodes remain internal.

## Not Mainnet

There is no real PHX value, no audited consensus, no staking enforcement, and no
production custody.
