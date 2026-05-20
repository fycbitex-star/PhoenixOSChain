# PhoenixChain v1.0 Mainnet Candidate Readiness

This document is a readiness checklist, not a launch approval. PhoenixChain must
not launch mainnet until audits, rehearsals, and governance approval are complete.

## Audit Checklist

Ready target:

- protocol specification
- threat model
- deterministic build instructions
- test coverage report
- fuzzing targets
- dependency review
- cryptography review
- consensus review
- storage recovery review

Placeholder:

- audit vendor selection
- remediation timeline

Acceptance criteria:

- all critical and high audit findings are fixed or explicitly risk-accepted

## Validator Requirements

Ready target:

- hardware requirements
- bandwidth requirements
- uptime expectations
- key management guide
- upgrade policy
- monitoring requirements

Placeholder:

- final validator agreement
- insurance or legal requirements

Acceptance criteria:

- every genesis validator confirms readiness with a signed checklist

## Treasury Multisig

Ready target:

- multisig participant list
- signing threshold
- emergency policy
- transparency process

Placeholder:

- final treasury allocation
- legal structure

Acceptance criteria:

- treasury transactions require documented multisig approval

## Genesis Allocation

Ready target:

- frozen allocation file
- reproducible genesis hash
- allocation rationale
- vesting placeholders if needed

Placeholder:

- final tokenomics approval

Acceptance criteria:

- genesis file is independently reproducible from published inputs

## Chain Parameters

Ready target:

- chain ID
- block time
- gas or fee rules
- validator set
- RPC limits
- upgrade rules

Placeholder:

- dynamic parameter governance

Acceptance criteria:

- chain parameters are documented and included in release artifacts

## Resilience

Ready target:

- backup RPC
- disaster recovery
- DDoS protection
- snapshot strategy
- restore procedure
- incident response runbook

Placeholder:

- regional infrastructure expansion

Acceptance criteria:

- disaster recovery is rehearsed before launch
- backup RPC endpoints are tested under load

## Launch Checklist

Ready target:

- release candidate tag
- signed binaries
- published genesis
- validator coordination window
- monitoring live
- status page live
- rollback decision process

Placeholder:

- final launch date

Acceptance criteria:

- launch is approved only after every checklist item is complete

## What Comes Next

After mainnet candidate readiness, the project either launches through an
approved process or returns to another candidate cycle.
