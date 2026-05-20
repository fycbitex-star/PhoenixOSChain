# PhoenixChain Smart Contracts

PhoenixChain smart contracts are currently supported through PCVM at dev/test
level.

## Implemented Example

`BasicTokenExample` stores an initial supply in contract storage slot `0` and
emits a log containing the supply.

## Transaction Types To Integrate Next

- deploy contract transaction
- call contract transaction

These are implemented in the VM package but are not yet part of block execution
in the live devnet.

## Receipts

Receipts include:

- contract address
- gas used
- success flag
- return value
- error
- logs

## What Comes Next

- wire VM execution into transaction processing
- persist receipts in storage indexes
- expose contract explorer endpoints
- design ABI and developer tooling
