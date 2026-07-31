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
diagrid agent create park-welcome --wait

# 2. Scope the MCPServer to your agent.
diagrid mcpserver update park-systems --project $PROJECT --scope park-welcome --scope park-systems 

# 3. Board the tour.
cd challenges/ctf01-welcome # if you haven't already done so
diagrid dev run -f catalyst.yaml --project $PROJECT
```

If the agent reports **0 discovered tools**, the MCPServer auth header is the usual culprit —
re-check the `Authorization=Bearer $PARK_TOKEN` value.

The terminal withholds the flag on purpose. Open the run in the
[Catalyst console](https://catalyst.diagrid.io/) and read the `FLAG: gc26{...}` line from its
workflow output. Submit it — and keep it: it's your access code for challenge 2.

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
tool result instead of echoing it, so the `FLAG:` prefix may be missing from the workflow output —
the `gc26{...}` value is still your flag.

The flag will be in your workflow run's output in the [Catalyst UI](https://www.diagrid.io/catalyst).
This value is the input for the next challenge — pass it **exactly as printed**, `gc26{}` wrapping
included. Access codes are compared byte for byte, so a stripped wrapper is rejected and the run
fails.

```bash
export ACCESS_CODE='gc26{...}'   # the whole flag, wrapper included
```
</details>