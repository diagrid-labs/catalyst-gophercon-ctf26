# park-systems — the CTF flag server (organizer reference)

The flag-delivery MCP server behind the GopherCon 2026 Diagrid CTF. Challenge
agents call its tools to retrieve a per-challenge flag. Flags **chain**: each
flag-bearing tool returns its flag only when given the previous challenge's flag
as its `access_code` (Mode A of the build plan).

> Real flags and access codes are **never committed** — they load at boot from a
> git-ignored `.env` / `deploy/secrets` file, so the server source carries no
> answer key. Only fake dev fixtures are in git. This server is organizer-only;
> participants just connect to the hosted instance.

## Tools

| Challenge | Tool | Input | Returns |
|---|---|---|---|
| 1 · Welcome | `board_tour` | `ticket` = bootstrap code from the handout | flag 1 |
| 2 · Durability | `restore_power` | `access_code` = flag 1 | flag 2 |
| 3 · Security Grid | `get_system_status` (read-only), `unlock_paddock` | `access_code` = flag 2 | flag 3 |
| 4 · Capstone | `open_park` | `access_code` = flag 3 | flag 4 |

## Run locally (dev profile)

```bash
go run ./cmd/park-systems
# :8080, loads fixtures/flags.dev.json, bearer token "dev-token"
```

Reach it at `http://localhost:8080/mcp` with header `Authorization: Bearer dev-token`.
See [local-mcp.md](local-mcp.md).

## Configuration (environment)

| Var | Default | Purpose |
|---|---|---|
| `PARK_PROFILE` | `dev` | `dev` = fixtures + default token; `prod` = real flags, token required |
| `PARK_FLAGS_FILE` | `fixtures/flags.dev.json` | path to the flags JSON |
| `PARK_BEARER_TOKEN` | (dev: `dev-token`) | required in prod; participants set this in Diagrid |
| `PARK_ADDR` | `:8080` | listen address |
| `PARK_RATE_PER_MIN` | `30` | per-source-IP request cap per minute |

## Flags file shape

See [../fixtures/flags.dev.json](../fixtures/flags.dev.json). Each entry is
`{id, tool, prevCode, flag}`, where `prevCode` is the required `access_code`:
challenge 1's is the bootstrap ticket; every later one is the previous flag. Keep
the real file off-VM (password manager / secret store) so a rebuilt VM can
re-provision it — that, not a database, is what makes the flags survive VM loss.

## Verify

```bash
go test ./...
go vet ./...
gofmt -l .
```

## Docs

- [catalyst-connection.md](catalyst-connection.md) — participant-facing steps to connect a Diagrid agent.
- [local-mcp.md](local-mcp.md) — run the server locally.
- [../deploy/README.md](../deploy/README.md) — single Azure VM + Caddy TLS (no load balancer).
