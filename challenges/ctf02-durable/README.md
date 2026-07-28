# Challenge 2 — Hold Onto Your Butts 🔌

**Category:** Catalyst · **Difficulty:** Medium · **Points:** 75

A durable agent (`power-grid`) restores the grid in steps: `reroute → breakers →
restore`. Kill it mid-run and start it again — Catalyst resumes from the last
completed step. The flag appears only **after the run survives the crash**. No
OpenAI key needed.

**You need:** `ACCESS_CODE` = your challenge 1 flag; the park-systems URL + token.

## Walkthrough

```bash
diagrid agent create power-grid --project $PROJECT --wait
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope power-grid \
  --header "Authorization=Bearer <PARK_TOKEN>"

go mod tidy
export ACCESS_CODE=<your challenge 1 flag>

# Prove it's durable: pin a run id, Ctrl-C during a step, then re-run the SAME command.
GOAI_RUN=butts-1 diagrid dev run -f catalyst.yaml --project $PROJECT
```

On the resumed run the final step calls `restore_power` and prints
`FLAG: gc26{...}` — your submission and the access code for challenge 3.

<details><summary>Run locally instead</summary>

```bash
export ACCESS_CODE=<challenge 1 flag>
dapr run --app-id power-grid --resources-path ./resources -- go run .
```
</details>
