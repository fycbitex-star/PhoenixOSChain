#!/usr/bin/env sh
set -eu

BACKUP_ROOT="${BACKUP_ROOT:-/root/backups}"
TS="$(date +%Y%m%d-%H%M%S)"
DEST="$BACKUP_ROOT/phoenixchain-runtime-$TS"
mkdir -p "$DEST"

tar -czf "$DEST/opt-phoenixchain.tar.gz" /opt/phoenixchain
tar -czf "$DEST/etc-phoenixchain.tar.gz" /etc/phoenixchain /etc/systemd/system/phoenixchain-node@.service /etc/nginx/sites-available/chain.fycbit.com /etc/nginx/sites-enabled/chain.fycbit.com
tar -czf "$DEST/var-lib-phoenixchain.tar.gz" /var/lib/phoenixchain

find "$BACKUP_ROOT" -maxdepth 1 -type d -name 'phoenixchain-runtime-*' | sort | head -n -7 | xargs -r rm -rf
echo "$DEST"
