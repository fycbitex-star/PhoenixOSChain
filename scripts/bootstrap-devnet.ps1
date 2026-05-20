$ErrorActionPreference = "Stop"

$env:Path = "C:\Program Files\Go\bin;$env:Path"
go run ./cmd/phoenixd devnet bootstrap --genesis configs/devnet.genesis.json --env .env.devnet --wallet-dir .phoenixchain/wallets
Write-Host "Load .env.devnet into your shell before running local non-Docker nodes."
