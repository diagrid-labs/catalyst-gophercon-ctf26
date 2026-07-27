# Running `park-systems` locally

For the Challenge 3 variant where participants run the MCP server themselves
instead of using the hosted one, and for local development of this repo.

## Run it

```bash
go run ./cmd/park-systems
# dev profile: :8080, fixtures/flags.dev.json, bearer token "dev-token"
```

With your own flags file and token:

```bash
PARK_FLAGS_FILE=./my-flags.json PARK_BEARER_TOKEN=my-token go run ./cmd/park-systems
```

## Point an MCP client at it

- Endpoint: `http://localhost:8080/mcp`
- Header:   `Authorization: Bearer dev-token`

## Register it in a Catalyst project

If your agent runs in Catalyst and can reach your machine (e.g. via a tunnel),
register the MCP server exactly as in
[catalyst-connection.md](catalyst-connection.md) but with your local/tunnel URL.

## Health check

```bash
curl -s localhost:8080/healthz   # -> ok
```
