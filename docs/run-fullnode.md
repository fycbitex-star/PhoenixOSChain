# Run A PhoenixChain Public Full Node

Full nodes can join the public testnet through the seed registry.

## Docker

```bash
docker compose -f docker-compose.devnet.yml up -d --build
```

## Linux

```bash
phoenixd start --config fullnode.yaml
```

## Seed Registration

```bash
curl -X POST https://chain.fycbit.com/seed/register \
  -H 'Content-Type: application/json' \
  -d '{"node_id":"my-node","address":"http://YOUR_PUBLIC_IP:18547"}'
```

## Seed Peers

```bash
curl https://chain.fycbit.com/seed/peers
```

## Limitations

The seed registry is basic and should be hardened before large public use.
