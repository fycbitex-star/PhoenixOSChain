# Governance Impact Receipts

PhoenixChain now carries a governance primitive that most governance systems still ignore:

not just "what was proposed",
but "what infrastructure surface would this decision actually touch if it ever moved forward".

## What It Is

A **Governance Impact Receipt** is a deterministic record derived from proposal runtime state.

Each receipt captures:

- proposal identity
- impact class
- execution criticality
- blast radius
- affected surfaces
- required preconditions
- active blockers
- deterministic digest

The goal is simple:

governance should not stop at narrative.
It should expose operational consequence.

## Why It Matters

Most governance systems expose:

- proposal text
- vote counts
- pass/fail posture

Very few expose:

- which trust surfaces are being touched
- what must become true before safe execution
- whether the decision is treasury-sensitive, validator-sensitive, or runtime-sensitive
- whether the proposal expands operational blast radius

PhoenixChain treats that missing layer as infrastructure debt.

## Why This Is Different

This repository does not present governance as forum theater.

It treats governance as a chain runtime that should eventually coordinate:

- treasury movement
- validator behavior
- runtime upgrade safety
- ecosystem policy
- operator trust

Impact receipts create a standard bridge between proposal text and infrastructure consequence.

## Current Posture

Impact receipts are currently:

- runtime-derived
- explorer-visible
- deterministic
- digest-backed
- intentionally conservative

They do **not** claim:

- live treasury execution
- live weighted quorum settlement
- live finality of execution

They instead standardize the governance consequence model while the chain is still in disciplined devnet form.

## Long-Term Value

If extended over time, Governance Impact Receipts can become the basis for:

- validator review workflows
- treasury risk gating
- governance simulation tooling
- AI governance analysis with better grounding
- proposal comparison by blast radius and criticality
- institutional governance policy review

That is the bigger thesis:

not just governance visibility,
but governance consequence legibility.
