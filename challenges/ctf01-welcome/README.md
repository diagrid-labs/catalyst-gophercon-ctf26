# Challenge 1 — Welcome to Jurassic Park 🦕

**Category:** Catalyst · **Difficulty:** Easy · **Points:** 50

Run your first Catalyst agent (`park-welcome`) and watch it appear under **Agents**.
It calls the park-systems `board_tour` tool with your entry ticket and hands back the flag.

**You need:** a model — [Ollama](https://ollama.com/download) (free, default) or an
`OPENAI_API_KEY` — your **entry ticket** (from the brief), and the **park-systems URL + bearer
token** (from the organizers). Set `PROJECT` to your Catalyst project.

## Walkthrough

```bash
# 0. Pick a model. Ollama is the free default — leave it running in its own terminal:
ollama serve
ollama pull qwen2.5:7b
# ...or use OpenAI instead, billed to your own account:
#   export OPENAI_API_KEY=sk-...

# 1. Register the agent — it shows up under Agents in the console.
diagrid agent create park-welcome --project $PROJECT --wait

# 2. Give the agent access to the park control systems. <PARK_TOKEN> is the
#    bearer token the organizers gave you — it authenticates every tool call.
#    (Confirm the exact auth-header flag with: diagrid mcpserver create --help)
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope park-welcome \
  --header "Authorization:Bearer <PARK_TOKEN>"

# 3. Board the tour.
go mod tidy
export PARK_TICKET=<your entry ticket>
diagrid dev run -f catalyst.yaml --project $PROJECT
```

If the agent reports **0 discovered tools**, the MCPServer auth header is the usual culprit —
re-check the `Authorization=Bearer <PARK_TOKEN>` value.

Find the `FLAG: gc26{...}` line in the output. Submit it — and keep it: it's your
access code for challenge 2.

<details><summary>Run locally instead</summary>

Run park-systems locally (dev profile: port 8080, bearer token `dev-token`), then:

```bash
export PARK_TICKET=<ticket>
dapr run --app-id park-welcome --resources-path ./resources -- go run .
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

Any model that supports **tool calling** works. Smaller local models sometimes paraphrase the
tool result instead of echoing it, so the `FLAG:` prefix may be missing — the `gc26{...}` value
is still your flag.
</details>
