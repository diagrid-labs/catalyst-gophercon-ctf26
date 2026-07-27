// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command ctf03-mcp is GopherCon CTF challenge 3 — "Access the Security Grid".
//
// A single-node LangChainGo agent that discovers the park-systems MCP tools,
// reads the grid with get_system_status, then calls unlock_paddock with your
// access code and reports the flag. It runs as a durable Dapr Workflow on
// Diagrid Catalyst and registers under Agents; the tool calls are visible in the
// Catalyst tool-call UI.
//
// The model is OpenAI (OPENAI_API_KEY, optionally OPENAI_MODEL, default gpt-4o).
// Set ACCESS_CODE to the flag you earned in challenge 2. See README.md.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	goai "github.com/diagridio/go-ai"
	"github.com/diagridio/go-ai/adapters/langchaingo"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/mcp"

	daprclient "github.com/dapr/go-sdk/client"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const agentName = "security-grid"

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
	fmt.Printf("Discovered %d MCP tool(s):\n", len(tools))
	for _, t := range tools {
		fmt.Printf("  - %s/%s: %s\n", t.Server, t.Name, t.Description)
	}

	graph := agent.NewGraph(agentName).
		AddNode("breach", langchaingo.ToolModelNode(newModel(), tools,
			langchaingo.WithSystemPrompt(
				"You are breaching the InGen security grid. First call get_system_status to read "+
					"the grid. Then call unlock_paddock exactly once, passing the provided access "+
					"code as the `access_code` argument. Reply with the unlock_paddock response "+
					"verbatim, including any line that starts with 'FLAG:'."),
			langchaingo.WithInputKey("input"),
			langchaingo.WithOutputKey("output"),
		)).
		SetEntry("breach").
		AddEdge("breach", agent.END)

	runner, err := goai.NewRunner(ctx, goai.Config{
		Graph:     graph,
		Name:      agentName,
		Framework: langchaingo.Framework,
		Role:      "Security Grid Operator",
		Goal:      "Access the security grid and disengage a paddock lock",
		Tools:     toolInfos(tools),
	})
	if err != nil {
		fatal(err)
	}
	defer runner.Close()

	accessCode := os.Getenv("ACCESS_CODE")
	if accessCode == "" {
		fatal(fmt.Errorf("set ACCESS_CODE to the flag you earned in challenge 2"))
	}

	out, err := runner.Invoke(ctx, state{Input: "The grid access code is: " + accessCode})
	if err != nil {
		fatal(err)
	}
	var result state
	if err := out.Into(&result); err != nil {
		fatal(err)
	}
	fmt.Printf("Agent %q ran as workflow %s\n", runner.Name(), runner.WorkflowName())
	fmt.Printf("Grid: %s\n", result.Output)
}

func newModel() llms.Model {
	if os.Getenv("OPENAI_API_KEY") == "" {
		fatal(fmt.Errorf("OPENAI_API_KEY is not set"))
	}
	m, err := openai.New(openai.WithModel(env("OPENAI_MODEL", "gpt-4o")))
	if err != nil {
		fatal(fmt.Errorf("openai model: %w", err))
	}
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

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
