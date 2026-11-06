# Deploying `park-systems` (single Azure VM)

Cheapest correct topology: **one small VM + a static public IP + Caddy for TLS.
No load balancer.** (An L4 Azure Load Balancer wouldn't terminate TLS, so Caddy
would stay anyway; Application Gateway would terminate TLS but costs far more than
a few-day event warrants.)

## Prerequisites
- A small Linux VM (B1s/B2s) with Docker + `docker compose`.
- A **Standard static public IP** attached to the VM's NIC.
- A **DNS A record** `park.<yourdomain>` → the public IP. Required: Let's Encrypt
  issues certs for names, not bare IPs.
- Inbound NSG: 443 from anywhere, 22 from your admin IP only. No other public ports.

## 1. Provision secrets (never committed)

```bash
cd deploy
cp .env.example .env                 # then edit: PARK_PROFILE=prod, a strong PARK_BEARER_TOKEN
mkdir -p secrets
# author secrets/flags.json with the REAL chain, same shape as ../fixtures/flags.dev.json
chmod 600 .env secrets/flags.json
```

Keep canonical copies of `.env` and `secrets/flags.json` in a password manager /
secret store so a rebuilt VM can re-provision them — that is your durability.

## 2. Set the domain

Edit `Caddyfile`, replace `park.example.com` with your DNS name.

## 3. Start

```bash
cd deploy
docker compose up -d --build
docker compose logs -f park-systems
```

Caddy obtains the TLS cert automatically. Verify:

```bash
curl -s https://park.<yourdomain>/healthz   # -> ok
```

Then confirm a Diagrid `MCPServer` resource can reach `https://park.<yourdomain>/mcp`.

## Ops

- **Rotate the bearer token:** edit `.env`, `docker compose up -d`.
- **Reload flags:** edit `secrets/flags.json`, `docker compose restart park-systems`.
- Logs are structured JSON and never contain flag values.
