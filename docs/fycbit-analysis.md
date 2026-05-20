# Fycbit.com Integration Analysis

Date: 2026-05-18

## Observed Infrastructure

- `fycbit.com` resolves to `45.84.191.49`.
- `chain.fycbit.com` resolves to `45.84.191.49`.
- Server OS: Ubuntu 24.04 LTS.
- Web server: nginx 1.24.0.
- Existing public app: Next.js behind nginx on `127.0.0.1:3040`.
- Existing API: PHP/Laravel behind nginx for `api.fycbit.com`.
- Existing EVM wallet app: node app behind nginx on `:3000`.
- Docker and certbot are installed on the server.

## Current Chain Subdomain State

- DNS is already pointed at the server.
- HTTP currently returns 404.
- HTTPS currently has a certificate mismatch because the existing certificate
  does not cover `chain.fycbit.com`.

## Integration Plan

PhoenixChain is installed as an isolated service:

- binary: `/opt/phoenixchain/bin/phoenixd`
- config: `/opt/phoenixchain/configs`
- data: `/var/lib/phoenixchain`
- environment: `/etc/phoenixchain/devnet.env`
- nginx site: `/etc/nginx/sites-available/chain.fycbit.com`
- public endpoint: `https://chain.fycbit.com`

The existing `fycbit.com`, `api.fycbit.com`, and `evm.fycbit.com` nginx configs
are backed up before changes. PhoenixChain only adds a new server block for
`chain.fycbit.com`.

## Site Improvement Notes

- The public HTML includes placeholder text such as `Last Updated: [Tarih]`.
- Favicon is configured as `YOUR FAVICON URL`.
- SEO title and description are generic and should be rewritten.
- Compliance content exists but should be legally reviewed before public claims.
- `chain.fycbit.com` should clearly identify itself as PhoenixChain Devnet v0.5,
  not production mainnet.

## Safety Notes

PhoenixChain Devnet v0.5 is not production-ready. It is suitable for local and
controlled server devnet testing only. No real customer deposits, trading, or
mainnet value should depend on it yet.
