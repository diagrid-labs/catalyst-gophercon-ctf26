# Challenge 1 — Welcome to Jurassic Park 🦕

**Category:** Catalyst · **Difficulty:** Easy · **Points:** 50

Run your first Catalyst agent (`park-welcome`) and watch it appear under **Agents**.
It calls the park-systems `board_tour` tool with your entry ticket and hands back the flag.

**You need:** `OPENAI_API_KEY`, your **entry ticket** (from the brief), and the
**park-systems URL + token**. Set `PROJECT` to your Catalyst project.

## Walkthrough

```bash
# 1. Register the agent — it shows up under Agents in the console.
diagrid agent create park-welcome --project $PROJECT --wait

# 2. Give the agent access to the park control systems.
#    (Confirm the exact auth-header flag with: diagrid mcpserver create --help)
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope park-welcome \
  --header "Authorization=Bearer <PARK_TOKEN>"

# 3. Board the tour.
go mod tidy
export OPENAI_API_KEY=sk-...  PARK_TICKET=<your entry ticket>
diagrid dev run -f catalyst.yaml --project $PROJECT
```

Find the `FLAG: gc26{...}` line in the output. Submit it — and keep it: it's your
access code for challenge 2.

<details><summary>Run locally instead</summary>

Run park-systems locally (dev profile: port 8080, token `dev-token`), then:

```bash
export OPENAI_API_KEY=sk-... PARK_TICKET=<ticket>
dapr run --app-id park-welcome --resources-path ./resources -- go run .
```
</details>
