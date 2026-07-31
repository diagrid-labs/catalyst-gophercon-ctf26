// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command ctf01-welcome is GopherCon CTF challenge 1 — "Welcome to Jurassic Park".
//
// A single-node LangChainGo agent that calls the park-systems MCP `board_tour`
// tool with your entry ticket and reports the flag it returns. It runs as a
// durable Dapr Workflow on Diagrid Catalyst and registers under Agents.
//
// The model runs locally on Ollama by default — no API key and no cost — or on
// OpenAI if you set OPENAI_API_KEY. Set PARK_TICKET to the entry ticket from the
// challenge brief. See README.md.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/internal/parkflag"
	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/internal/parkmodel"

	goai "github.com/diagridio/go-ai"
	"github.com/diagridio/go-ai/adapters/langchaingo"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/mcp"

	daprclient "github.com/dapr/go-sdk/client"
	"github.com/tmc/langchaingo/llms"
)

const agentName = "park-welcome"

type state struct {
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

func main() {
	ctx := context.Background()

	dc, err := daprclient.NewClient()
	if err != nil {
		fatal(fmt.Errorf("dapr client: %w", err))
	}
	defer dc.Close()
	wf, err := daprclient.NewWorkflowClient()
	if err != nil {
		fatal(fmt.Errorf("workflow client: %w", err))
	}

	tools, err := mcp.DiscoverTools(ctx, dc, wf)
	if err != nil {
		fatal(fmt.Errorf("discover MCP tools: %w", err))
	}
	fmt.Printf("Discovered %d MCP tool(s) from the sidecar.\n", len(tools))
	// With no tools bound, the model has no way to reach park-systems and will
	// happily invent a flag instead. Stop before it can.
	if len(tools) == 0 {
		fatal(fmt.Errorf("no MCP tools discovered: the agent cannot reach park-systems.\n"+
			"  Check the MCPServer is scoped to this app ID (%s) and that its\n"+
			"  Authorization header is exactly \"Bearer <token>\" — no colon after Bearer", agentName))
	}

	graph := agent.NewGraph(agentName).
		AddNode("gate", langchaingo.ToolModelNode(newModel(), tools,
			langchaingo.WithSystemPrompt(
				"You are the gate to Jurassic Park. Call the board_tour tool exactly once, "+
					"passing the visitor's ticket as the `ticket` argument. Then reply with the "+
					"tool's response copied verbatim, character for character, including the "+
					"line that starts with 'FLAG:'. Never write a 'FLAG:' line the tool did "+
					"not return, and never paraphrase, summarize, or add welcome text of your "+
					"own. If the tool returned no response, say exactly that."),
			langchaingo.WithInputKey("input"),
			langchaingo.WithOutputKey("output"),
		)).
		SetEntry("gate").
		AddEdge("gate", agent.END)

	runner, err := goai.NewRunner(ctx, goai.Config{
		Graph:     graph,
		Name:      agentName,
		Framework: langchaingo.Framework,
		Role:      "Park Gate",
		Goal:      "Welcome visitors and board them onto the tour",
		Tools:     toolInfos(tools),
	})
	if err != nil {
		fatal(err)
	}
	defer runner.Close()

	ticket := os.Getenv("PARK_TICKET")
	if ticket == "" {
		fatal(fmt.Errorf("set PARK_TICKET to your entry ticket (from the challenge brief)"))
	}

	out, err := runner.Invoke(ctx, state{Input: "My ticket is: " + ticket})
	if err != nil {
		fatal(err)
	}
	var result state
	if err := out.Into(&result); err != nil {
		fatal(err)
	}
	fmt.Printf("Agent %q ran as workflow %s\n", runner.Name(), runner.WorkflowName())
	fmt.Printf("Gate: %s\n", parkflag.Redact(result.Output))
}

// newModel builds the chat model. Provider selection, defaults and the Ollama
// preflight all live in the shared parkmodel package, so every model-driven
// challenge behaves identically.
func newModel() llms.Model {
	m, err := parkmodel.New()
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Model: %s\n", m)
	return m
}

// toolInfos declares the discovered tools for the Catalyst Agents view.
func toolInfos(tools []mcp.Tool) []goai.ToolInfo {
	out := make([]goai.ToolInfo, len(tools))
	for i, t := range tools {
		args, _ := json.Marshal(t.Schema)
		out[i] = goai.ToolInfo{Name: t.Name, Description: t.Description, Args: string(args)}
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
