# PhoenixChain Validator Roadmap

PhoenixChain validators must be introduced gradually. v0.1 and v0.5 validator
logic is for controlled development only.

## Validator Registration

Ready target:

- local config-based validator list
- public testnet registration process
- validator metadata format

Placeholder:

- on-chain validator registry
- staking-based admission

Acceptance criteria:

- devnet validators are clearly configured and auditable

## Validator Identity

Ready target:

- node identity
- validator address
- signing public key
- operator metadata

Placeholder:

- decentralized identity
- reputation system

Acceptance criteria:

- each validator can be uniquely identified in logs and explorer views

## Validator Keys

Ready target:

- separate validator signing key
- no hardcoded keys
- local encrypted storage
- rotation instructions

Placeholder:

- remote signer
- hardware security module

Acceptance criteria:

- validator keys are generated locally and excluded from source control

## Validator Rotation

Ready target:

- manual rotation in devnet
- documented rotation procedure

Placeholder:

- automatic on-chain rotation
- governance-controlled validator changes

Acceptance criteria:

- a validator can be replaced in a controlled devnet upgrade

## Staking Placeholder

Ready target:

- staking design document
- staking state model
- delegation considerations

Placeholder:

- real staking transactions
- reward economics

Acceptance criteria:

- staking design is reviewed before implementation

## Slashing Placeholder

Ready target:

- slashing risk model
- double-sign detection plan
- downtime policy draft

Placeholder:

- active slashing enforcement

Acceptance criteria:

- slashing rules are not implemented until they are specified and reviewed

## Uptime Monitoring

Ready target:

- validator block signing status
- peer count
- height lag
- missed block count

Placeholder:

- automated penalties

Acceptance criteria:

- operators can see validator health in testnet dashboards

## Validator Rewards Placeholder

Ready target:

- reward accounting design
- distribution schedule draft

Placeholder:

- real reward distribution

Acceptance criteria:

- rewards are not activated before tokenomics is finalized

## What Comes Next

The next validator task is v0.5 local PoA validator networking and block
propagation. Staking and slashing remain future work.
