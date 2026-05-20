# Publish Checklist

Use this checklist when turning PhoenixOSChain into a public GitHub repository.

## Repository Settings

1. repository name: `PhoenixOSChain`
2. visibility: `Public`
3. do not initialize with a template README, gitignore, or license
4. enable branch protection for `main`
5. require pull requests for protected branches

## Before First Push

1. verify no secrets or runtime state exist in the working tree
2. verify `LICENSE`, `NOTICE`, and `SECURITY.md` are present
3. verify public docs are accurate and do not overclaim production readiness
4. run `go test ./...` in a Go-enabled environment

## First Push

```powershell
git remote add origin https://github.com/<owner>/PhoenixOSChain.git
git push -u origin main
```

## After Push

1. pin the key architecture docs in the README
2. enable security alerts
3. enable secret scanning if available
4. add repository topics:
   - blockchain
   - layer1
   - validator
   - governance
   - treasury
   - evm
   - fintech
   - infrastructure

## Messaging Rule

Public positioning should sound like institutional infrastructure:

- precise
- high-conviction
- technically grounded
- never fake
