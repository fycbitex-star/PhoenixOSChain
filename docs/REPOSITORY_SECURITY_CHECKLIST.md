# Repository Security Checklist

This checklist is for keeping PhoenixOSChain public without turning it into an operational leak surface.

## 1. GitHub Repository Settings

- set default branch protection on `main`
- require pull requests before merge
- require at least one review for protected branches
- block force-push on protected branches
- block branch deletion on protected branches
- enable secret scanning if available
- enable dependency alerts if available
- enable private vulnerability reporting path or direct security contact

## 2. Public vs Private Boundary

Public is allowed:

- source code
- public architecture docs
- validator model direction
- governance model direction
- treasury safety philosophy
- devnet and testnet operational docs without secrets

Private only:

- validator private keys
- signer material
- treasury signer identity and signing procedures
- seed phrases
- production `.env` files
- internal hostnames and internal IP ranges
- backup bucket names and restore credentials
- webhook secrets
- RPC origin allowlists
- recovery command bundles with live credentials

## 3. Commit Hygiene

- never commit `.env`
- never commit wallet dumps
- never commit keystore passwords
- never commit private chain state snapshots that contain operational secrets
- never commit copied SSH sessions or terminal dumps with auth material
- avoid screenshots that expose tokens, hostnames, or operator emails

## 4. Release Hygiene

Before every public release:

- run a manual scan for `private key`, `seed`, `token`, `secret`, `passphrase`, `ssh`, `bearer`
- verify no devnet bootstrap output with private material is tracked
- verify no recovery archive contains live operational secrets
- verify docs do not expose internal topology
- verify public claims do not exceed audited reality

## 5. Branch Protection Discipline

- use `main` for public history only
- keep internal ops branches outside the public repo when they include sensitive deployment detail
- do not use the public repository as a live ops notebook
- prefer curated docs over raw operator transcripts

## 6. Issue and Discussion Safety

- disable blank issues
- use templates that warn against posting secrets
- redirect vulnerability reports away from public issues
- remove any issue that leaks credentials or internal topology immediately

## 7. CI and Automation

- CI may run tests and builds
- CI must not expose production credentials in logs
- CI must not print secret environment variables
- CI must use least-privilege tokens only
- release workflows should not contain deploy credentials for unrelated infrastructure

## 8. Artifact Policy

Allowed:

- source archives
- docs bundles
- public release notes

Not allowed:

- production database snapshots
- raw validator environment bundles
- signed treasury transaction payloads
- operator password dumps
- internal recovery tarballs with sensitive metadata

## 9. Security Disclosure Policy

- public repo is for architecture and code visibility
- vulnerability disclosure is private
- exploit paths are not discussed publicly until fixed
- postmortems must be redacted for operationally sensitive detail

## 10. Final Rule

PhoenixOSChain should look transparent without becoming operationally porous.

Public clarity is the goal.
Operational leakage is not.
