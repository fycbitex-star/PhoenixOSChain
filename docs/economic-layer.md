# PhoenixChain Economic Layer

The economic layer is functional at dev/test level only. It is not final
tokenomics.

## Locked Native Asset Reference

- asset: `PHX`
- decimals: `18`
- maximum supply: `30,000,000 PHX`
- compact display: `30M PHX`

This reference is canonical even though staking, validator rewards, treasury
flows and governance execution are not yet live.

## Implemented

- configurable economic parameters
- validator registration
- validator metadata
- validator bonding
- unbonding state
- validator reward accounting
- treasury fee accounting
- slashing placeholder event logic

## Not Final

- staking economics
- inflation
- validator reward schedule
- treasury policy
- slashing rules
- governance

## Next Steps

- publish tokenomics draft
- simulate validator economics
- connect fees from transaction execution
- design governance-controlled parameter updates
