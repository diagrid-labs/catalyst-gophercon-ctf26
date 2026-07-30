# Connecting your Catalyst agent to `park-systems` (MCP)

These steps register the `park-systems` MCP server in your Catalyst project and
wire it into your agent graph so your agent can call the challenge tools.

> The event organizers give you both values before you start:
> `<PARK_URL>` = e.g. `https://park.<domain>/mcp` · `<PARK_TOKEN>` = the event **bearer token**.
> This page contains **no flags** — you earn those by calling the tools.

## 0. The bearer token

park-systems accepts nothing unauthenticated. Every request must carry:

```
Authorization: Bearer <PARK_TOKEN>
```

You never put that token in your agent code. It lives on the **MCPServer
resource** in Catalyst, which attaches the header to each tool call your agent
makes. One token is shared by all participants for the whole event, and it works
for every challenge — register it once per agent and you're done.

If you're running park-systems yourself instead of using the hosted one, the
token is whatever `PARK_BEARER_TOKEN` is set to (`dev-token` in the dev profile)
— see [local-mcp.md](local-mcp.md).

## 1. Register the MCP server in Catalyst

With the `diagrid` CLI (confirm exact flags with `diagrid mcpserver create --help`):

```bash
diagrid mcpserver create park-systems \
  --project <your-project> \
  --url <PARK_URL> \
  --transport streamable-http \
  --scope <your-agent> \
  --header "Authorization=Bearer <PARK_TOKEN>"
```

Or create an **MCPServer** resource in the Catalyst UI: Streamable HTTP endpoint
= `<PARK_URL>`, with an `Authorization: Bearer <PARK_TOKEN>` header.

Confirm it is Ready:

```bash
diagrid mcpserver get park-systems --project <your-project>
```

**Check the token first when things break.** A wrong or missing token surfaces as
`401 Unauthorized`, or as your agent discovering **0 tools** — not as an obvious
auth error in your own code.

## 2. Add an MCP tool node to your agent graph

In your `go-ai` agent graph, add an MCP tool node bound to the `park-systems`
server. Your agent can now discover and call its tools.

## 3. Call the tool for your challenge

| Challenge | Call | Argument |
|---|---|---|
| 1 | `board_tour` | `ticket`: your entry ticket from the challenge brief |
| 2 | `restore_power` | `access_code`: the flag from challenge 1 |
| 3 | `get_system_status`, then `unlock_paddock` | `access_code`: the flag from challenge 2 |
| 4 | `open_park` | `access_code`: the flag from challenge 3 |

`get_system_status` is a safe, read-only probe — use it to see a tool call render
in the Catalyst tool-call UI before you attempt `unlock_paddock`.

## 4. Read the flag

The tool result is a text block. On success it contains a line:

```
FLAG: gc26{...}
```

That flag is **both** your Challenge Series submission **and** the `access_code`
for the next challenge — the challenges build on each other.

If you see **"Access code rejected"**, the `access_code` you passed doesn't match.
Make sure you completed the previous challenge and passed its exact flag.
