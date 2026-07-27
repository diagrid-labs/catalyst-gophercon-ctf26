# 🦖 GopherCon 2026 CTF — Jurassic Park on Diagrid Catalyst

Welcome! Bring Jurassic Park back online by building **durable AI agents in Go**
that run on [Diagrid Catalyst](https://www.diagrid.io/catalyst). In each
challenge your agent talks to the park's control systems and returns a **flag**
(`gc26{...}`) — submit it to score, then use it as the key to the next challenge.

## Before you start

- A [Catalyst account + the `diagrid` CLI](https://docs.diagrid.io/catalyst/references/cli-reference/), and Go 1.26.4+.
- An `OPENAI_API_KEY` for the AI challenges (1 and 3).
- Your **entry ticket** and the **park-systems connection details** — from the event brief.
- The challenges build on the private [go-ai](https://github.com/diagridio/go-ai) framework, so before `go mod tidy` / `go run`:
  ```bash
  export GOPRIVATE=github.com/diagridio/*
  ```
- One-time Catalyst setup:
  ```bash
  diagrid login
  diagrid project update <your-project> --enable-agent-infrastructure
  ```

## The challenges

Start at #1 and follow the flags — each flag you earn is the access code for the next.

| # | Challenge | You'll learn |
|---|---|---|
| 1 | [Welcome to Jurassic Park](challenges/ctf01-welcome/README.md) 🦕 | run your first agent, see it in Catalyst |
| 2 | [Hold Onto Your Butts](challenges/ctf02-durable/README.md) 🔌 | crash a durable workflow and watch it resume |
| 3 | [Access the Security Grid](challenges/ctf03-mcp/README.md) 🔐 | register an MCP server and call its tools |
| 4 | [Life Finds a Way](challenges/ctf04-capstone/README.md) 🦖 | run several agents together |

## The park control system (park-systems)

Your agents fetch their flags from **park-systems**, our hosted MCP "park control
system," running on the event VM at **[TODO — park-systems URL](https://TODO)**.
You don't run it — the organizers share the connection details (URL + access
token) with you before you start.

The park-systems **server source is committed to this repo only *after* the
challenge**, so the flag values and unlock logic aren't exposed ahead of time.

Hold onto your butts. 🦕
