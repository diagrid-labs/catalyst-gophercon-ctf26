# Challenge 3 — Access the Security Grid 🔐

**Category:** Catalyst · **Difficulty:** Medium · **Points:** 75

Register an MCP server in Catalyst and let your agent (`security-grid`) call its
tools. The agent reads the grid with `get_system_status`, then calls
`unlock_paddock` with your access code and returns the flag — and you can watch
both tool calls in the Catalyst tool-call UI.

**You need:** `OPENAI_API_KEY`; `ACCESS_CODE` = your challenge 2 flag; the
park-systems URL + token.

## Walkthrough

```bash
diagrid agent create security-grid --project $PROJECT --wait

# The heart of this challenge: register park-systems as a governed MCP server,
# scoped to this agent.
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope security-grid \
  --header "Authorization:Bearer <PARK_TOKEN>"

go mod tidy
export OPENAI_API_KEY=sk-...  ACCESS_CODE=<your challenge 2 flag>
diagrid dev run -f catalyst.yaml --project $PROJECT
```

The agent lists the discovered tools and prints `FLAG: gc26{...}`. Open the run
in the Catalyst UI to see the `get_system_status` and `unlock_paddock` calls. The
flag is your access code for challenge 4.

<details><summary>Run locally instead</summary>

```bash
export OPENAI_API_KEY=sk-... ACCESS_CODE=<challenge 2 flag>
dapr run --app-id security-grid --resources-path ./resources -- go run .
```
</details>
