# Challenge 4 — Life Finds a Way 🦖

**Category:** Catalyst · **Difficulty:** Hard · **Points:** 100

The finale: **three** agents run together and each appears as its own entry under
**Agents**. `raptor-keeper` and `grid-ranger` bring their systems online;
`park-director` opens the park by calling `open_park` with your access code and
returns the final flag. All three agents are deterministic — no model, so you
don't even need Ollama running.

**You need:** `ACCESS_CODE` = your challenge 3 flag; the park-systems URL +
**bearer token**.

## Walkthrough

```bash
# An Agent resource for each — all three show up in the Agents tab.
diagrid agent create raptor-keeper --project $PROJECT --wait
diagrid agent create grid-ranger   --project $PROJECT --wait
diagrid agent create park-director --project $PROJECT --wait

# park-director is the one that opens the park.
diagrid mcpserver create park-systems --project $PROJECT \
  --url <PARK_URL> --transport streamable-http --scope park-director \
  --header "Authorization=Bearer <PARK_TOKEN>"

go mod tidy
export ACCESS_CODE=<your challenge 3 flag>
diagrid dev run -f catalyst.yaml --project $PROJECT   # starts all three agents
```

`park-director` prints the final `FLAG: gc26{...}`. Welcome... to Jurassic Park.

<details><summary>Run locally instead</summary>

Three terminals (one App ID each):

```bash
dapr run --app-id raptor-keeper --resources-path ./resources -- go run ./cmd/raptor-keeper
dapr run --app-id grid-ranger   --resources-path ./resources -- go run ./cmd/grid-ranger
ACCESS_CODE=<challenge 3 flag> dapr run --app-id park-director --resources-path ./resources -- go run ./cmd/park-director
```
</details>
