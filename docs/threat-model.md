# PhoenixChain Threat Model

## Assets

- validator private keys
- user wallet keys
- chain state
- RPC availability
- block integrity
- transaction integrity

## Threats

- stolen validator keys
- malformed transactions
- block forgery
- RPC flooding
- peer eclipse attacks
- database corruption
- replay attacks
- wallet keystore theft

## Existing Mitigations

- signatures
- nonce checks
- validator authorization
- block hash validation
- RPC rate limiting
- encrypted wallet keystore
- Badger snapshot validation

## Open Risks

- no BFT finality
- no slashing
- limited peer scoring
- no hardware key support
- no formal audit
