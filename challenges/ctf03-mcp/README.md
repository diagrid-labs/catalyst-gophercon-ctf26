# Challenge 3 — Access the Security Grid 🔐

**Category:** Catalyst · **Difficulty:** Medium · **Points:** 75

Register an MCP server in Catalyst and let your agent (`security-grid`) call its
tools. The agent reads the grid with `get_system_status`, then calls
`unlock_paddock` with your access code and returns the flag — and you can watch
both tool calls in the Catalyst tool-call UI.

**You need:** a model — [Ollama](https://ollama.com/download) (free, default) or an
`OPENAI_API_KEY`; `ACCESS_CODE` = your challenge 2 flag; the park-systems URL + **bearer token**.

## Walkthrough

```bash
# Pick a model. Ollama is the free default — leave it running in its own terminal:
ollama serve
ollama pull qwen2.5:7b
# ...or use OpenAI instead, billed to your own account:
#   export OPENAI_API_KEY=sk-...

diagrid agent create security-grid --project $PROJECT --wait

# The heart of this challenge: register park-systems as a governed MCP server,
# scoped to this agent. The bearer token from the organizers goes in the auth
# header — Catalyst attaches it to every tool call your agent makes.
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope security-grid \
  --header "Authorization:Bearer <PARK_TOKEN>"

go mod tidy
export ACCESS_CODE=<your challenge 2 flag>
diagrid dev run -f catalyst.yaml --project $PROJECT
```

This challenge prints the discovered tools first — an empty list means the bearer token or the
auth header is wrong, not that your agent is broken.

The agent lists the discovered tools and prints `FLAG: gc26{...}`. Open the run
in the Catalyst UI to see the `get_system_status` and `unlock_paddock` calls. The
flag is your access code for challenge 4.

<details><summary>Run locally instead</summary>

```bash
export ACCESS_CODE=<challenge 2 flag>
dapr run --app-id security-grid --resources-path ./resources -- go run .
```
</details>

<details><summary>Choosing a model / provider</summary>

The agent prints the model it's using at startup. Defaults to Ollama; uses OpenAI if
`OPENAI_API_KEY` is set. The selection logic is shared by every model-driven challenge and lives
in [`challenges/internal/parkmodel`](../internal/parkmodel/).

| Variable | Effect |
|---|---|
| `MODEL_PROVIDER` | `ollama` or `openai` — forces one, overriding auto-detection |
| `OLLAMA_MODEL` | Ollama model (default `qwen2.5:7b`) |
| `OLLAMA_BASE_URL` | Ollama endpoint (default `http://localhost:11434/v1`) |
| `OPENAI_API_KEY` | your OpenAI key — its presence selects OpenAI |
| `OPENAI_MODEL` | OpenAI model (default `gpt-4o`) |
| `OPENAI_BASE_URL` | optional, for an OpenAI-compatible gateway |

This challenge needs **two** tool calls in sequence, so pick a model you've seen handle
multi-step tool use — `qwen2.5:7b` and up are reliable here, as is `gpt-4o`.
</details>
