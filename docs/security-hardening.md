# PhoenixChain v0.6 Security Hardening

PhoenixChain v0.6 is a hardened devnet, not production mainnet.

## Implemented Server Controls

- `chain.fycbit.com` is isolated behind nginx reverse proxy.
- Public traffic reaches only node-3 on localhost port `18547`.
- Validator nodes listen on localhost ports `18545` and `18546`.
- nginx rate limiting is configured for PhoenixChain RPC.
- nginx request body size is limited to 1 MB.
- PhoenixChain RPC has application-level request size limits.
- PhoenixChain RPC has per-client rate limiting.
- PhoenixChain RPC has read/write/idle timeouts.
- Structured JSON logs are emitted by nodes.
- systemd restarts failed node services.

## SSH Hardening Plan

Required controls:

- create a non-root sudo admin user
- install SSH public key authentication
- verify key login before disabling password login
- disable password SSH login
- disable direct root SSH login when recovery access is confirmed

These steps must be performed carefully to avoid lockout.

## Validator Key Strategy

- Validator keys are stored in `/etc/phoenixchain/devnet.env`.
- File mode should be `0600`.
- Keys are local devnet keys only.
- Keys must not be committed to git.
- Future work: encrypted keystore or remote signer.

## Remaining Security Work

- hardware-backed validator signing
- network-level DDoS protection
- API key support for privileged endpoints
- public testnet threat model
- external audit before any mainnet candidate
