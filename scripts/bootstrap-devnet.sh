#!/usr/bin/env sh
set -eu

go run ./cmd/phoenixd devnet bootstrap --genesis configs/devnet.genesis.json --env .env.devnet --wallet-dir .phoenixchain/wallets
echo "Load .env.devnet into your shell before running local non-Docker nodes."
