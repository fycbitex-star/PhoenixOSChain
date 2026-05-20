# PhoenixChain Disaster Recovery

PhoenixChain v0.6 includes backup and restore scripts for devnet operations.

## Backup

Run on server:

```bash
/opt/phoenixchain/scripts/backup-phoenixchain.sh
```

Backups include:

- `/opt/phoenixchain`
- `/etc/phoenixchain`
- systemd service file
- nginx chain site config
- `/var/lib/phoenixchain`

## Restore

Run:

```bash
/opt/phoenixchain/scripts/restore-phoenixchain.sh /root/backups/phoenixchain-runtime-YYYYmmdd-HHMMSS
```

## Emergency Restart

```bash
systemctl restart phoenixchain-node@1 phoenixchain-node@2 phoenixchain-node@3
systemctl status phoenixchain-node@1 phoenixchain-node@2 phoenixchain-node@3
```

## Rollback

Use the latest backup from `/root/backups` or the previous `/opt/phoenixchain`
directory created during deployment.

## Remaining Work

- off-server backups
- encrypted backup storage
- restore drills
- public testnet incident response rota
