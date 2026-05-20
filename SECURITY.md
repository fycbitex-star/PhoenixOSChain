# Security Posture

## Current Truth

PhoenixOSChain is not yet mainnet-ready software.

The repository contains meaningful chain, validator, governance, and explorer code, but public launch requires additional work in:

- consensus hardening
- treasury and governance mutation safety
- validator decentralization
- key management discipline
- disaster recovery rehearsal
- external security review

## Reporting

Do not publish security findings in public issues.

Report critical findings privately to the repository owner/operator team.

## Current Security Focus

1. validator and operator key hygiene
2. governance and treasury execution safety
3. RPC and public-surface abuse resistance
4. indexer and explorer integrity
5. mainnet launch discipline

## Public Disclosure Boundary

The public repository must not be used to disclose:

- validator private keys
- signer material
- treasury signing paths
- seed phrases
- internal IP ranges
- recovery bucket names or restore credentials
- production webhook secrets
- authentication tokens
- private incident procedures that expose live attack paths

Public publication should explain the system clearly without turning the repository into an operator compromise guide.

## Public Materials Rule

Architectural clarity is public.
Operational secrets are not.

## Operator Checklist

Use [docs/REPOSITORY_SECURITY_CHECKLIST.md](docs/REPOSITORY_SECURITY_CHECKLIST.md) before every public push, release, and documentation update.
