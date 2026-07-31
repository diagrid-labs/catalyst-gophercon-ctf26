# Challenge 2 — Hold Onto Your Butts 🔌

**Category:** Catalyst · **Difficulty:** Medium · **Points:** 75

A durable agent (`power-grid`) restores the grid in steps: `reroute → breakers →
restore`. Kill it mid-run and start it again — Catalyst resumes from the last
completed step. The flag appears only **after the run survives the crash**. This
agent is deterministic — no model, so you don't even need Ollama running.

**You need:** `ACCESS_CODE` = your challenge 1 flag; the park-systems URL +
**bearer token**.

## Walkthrough

Create this challenge's agent and scope the MCPServer to allow the agent access to it.

```bash
diagrid agent create power-grid --project $PROJECT --wait
diagrid mcpserver update park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope power-grid \
  --header "Authorization:Bearer <PARK_TOKEN>"

export ACCESS_CODE='gc26{...}' # whole challenge one flag, wrapper included

# Trip the grid on purpose, once `reroute` has been checkpointed.
GOAI_RUN=butts-1 FAIL_AT=breakers diagrid dev run -f catalyst.yaml --project $PROJECT
```

`FAIL_AT=<node>` fails that node once. Nodes are `reroute`, `breakers`, `restore`.
Nothing to interrupt, nothing to re-run — just let it ride.

Watch the output:

```
... EXECUTING reroute        ← runs, gets checkpointed
=== FAIL_AT=breakers ...     ← trips, 5s pause
... retrying breakers        ← retry comes due
... EXECUTING breakers       ← no second "reroute" line
```

**That missing `reroute` line is the whole point.** Its result was replayed from
the workflow history instead of being recomputed. Every node is checkpointed, so
only unfinished work runs again.

Once the retry clears, the final step calls `restore_power`. The terminal withholds
the flag — read the `FLAG: gc26{...}` line from the run's workflow output in the
[Catalyst console](https://catalyst.diagrid.io/). It's your submission and the
access code for challenge 3.

<details><summary>Run locally instead</summary>

```bash
dapr run --app-id power-grid --resources-path ./resources -- go run .
```

The flag will be in your workflow run's output in the [Catalyst UI](https://www.diagrid.io/catalyst).
This value is the input for the next challenge — pass it **exactly as printed**, `gc26{}` wrapping
included. Access codes are compared byte for byte, so a stripped wrapper is rejected and the run
fails.

```bash
export ACCESS_CODE='gc26{...}'   # the whole flag, wrapper included
```
</details>
