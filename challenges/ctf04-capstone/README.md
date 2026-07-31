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
diagrid agent create grid-ranger --project $PROJECT --wait
diagrid agent create park-director --project $PROJECT --wait

# park-director is the one that opens the park.
diagrid mcpserver update park-systems --scope park-director --project $PROJECT 

export ACCESS_CODE='gc26{...}' # whole challenge three flag, wrapper included
diagrid dev run -f catalyst.yaml --project $PROJECT   # starts all three agents
```

`park-director` withholds the flag in the terminal. Read the final `FLAG: gc26{...}`
from its workflow output in the [Catalyst console](https://catalyst.diagrid.io/).
Welcome... to Jurassic Park.

<details><summary>Alternatively:</summary>

Three terminals (one App ID each):

```bash
dapr run --app-id raptor-keeper --resources-path ./resources -- go run ./cmd/raptor-keeper
dapr run --app-id grid-ranger   --resources-path ./resources -- go run ./cmd/grid-ranger
ACCESS_CODE=<challenge 3 flag> dapr run --app-id park-director --resources-path ./resources -- go run ./cmd/park-director
```
</details>
