# Connecting your Catalyst agent to `park-systems` (MCP)

These steps register the `park-systems` MCP server in your Catalyst project and
wire it into your agent graph so your agent can call the challenge tools.

> The event organizers give you the server URL and access token.
> `<PARK_URL>` = e.g. `https://park.<domain>/mcp` · `<PARK_TOKEN>` = the event bearer token.
> This page contains **no flags** — you earn those by calling the tools.

## 1. Register the MCP server in Catalyst

With the `diagrid` CLI (confirm exact flags with `diagrid mcpserver create --help`):

```bash
diagrid mcpserver create park-systems \
  --project <your-project> \
  --url <PARK_URL> \
  --header "Authorization:Bearer <PARK_TOKEN>"
```

Or create an **MCPServer** resource in the Catalyst UI: Streamable HTTP endpoint
= `<PARK_URL>`, with an `Authorization: Bearer <PARK_TOKEN>` header.

Confirm it is Ready:

```bash
diagrid mcpserver get park-systems --project <your-project>
```

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
