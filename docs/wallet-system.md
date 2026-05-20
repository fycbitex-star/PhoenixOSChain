# Phoenix Wallet System

Phoenix Wallet is the account and signing layer for PhoenixChain.

## Implemented

- Ed25519 wallet generation
- address derivation
- encrypted local keystore files
- wallet load/decrypt
- transaction signing API
- mnemonic placeholder generation

## Keystore

The keystore uses:

- AES-GCM encryption
- password-derived key
- per-file salt
- authenticated address data

## Limitations

- mnemonic is not BIP39 yet
- no hardware wallet support
- no browser extension yet
- no production custody

## Next Steps

- replace mnemonic placeholder with reviewed standard
- add CLI keystore commands
- add wallet RPC integration
- design browser extension architecture
