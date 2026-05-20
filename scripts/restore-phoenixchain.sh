#!/usr/bin/env sh
set -eu

if [ "${1:-}" = "" ]; then
  echo "usage: restore-phoenixchain.sh /root/backups/phoenixchain-runtime-YYYYmmdd-HHMMSS" >&2
  exit 1
fi

SRC="$1"
systemctl stop phoenixchain-node@1 phoenixchain-node@2 phoenixchain-node@3 2>/dev/null || true
tar -xzf "$SRC/opt-phoenixchain.tar.gz" -C /
tar -xzf "$SRC/etc-phoenixchain.tar.gz" -C /
tar -xzf "$SRC/var-lib-phoenixchain.tar.gz" -C /
systemctl daemon-reload
nginx -t
systemctl reload nginx
systemctl start phoenixchain-node@1 phoenixchain-node@2 phoenixchain-node@3
echo "PhoenixChain restored from $SRC"
