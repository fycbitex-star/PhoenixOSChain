# PHX-20 Standard

`PHX-20` is the official fungible token foundation for PhoenixChain.

Current scope:
- `name`
- `symbol`
- `decimals`
- `totalSupply`
- `balanceOf`
- `transfer`
- `approve`
- `allowance`
- `transferFrom`
- optional `mint`
- optional `burn`
- optional `owner/admin`
- optional `pause`

Required events:
- `Transfer`
- `Approval`
- `Mint`
- `Burn`
- `Paused`
- `Unpaused`

Current runtime truth:
- live as a `PCVM devnet token foundation`
- token metadata, holders, balances and transfer events are real runtime records
- public chain transaction calldata plumbing is still maturing
- wallet-signed browser token transfers are not live yet
- this is not mainnet and not audited

Templates:
- `PHX20 Basic`
- `PHX20 Mintable`
- `PHX20 Burnable`
- `PHX20 Pausable`
- `PHX20 Governance-ready`

Explorer surfaces:
- `/tokens`
- `/tokens/{tokenAddress}`
- `/tokens/factory`
- `/wallet/assets/{address}`
- `/api/phx20/spec`
