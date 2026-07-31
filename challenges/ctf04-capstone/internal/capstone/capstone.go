// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package capstone holds helpers shared by the challenge 4 agents. Each agent is
// a separate program (its own App ID / Agent resource) so they all appear as
// distinct entries under Agents in the Catalyst console.
package capstone

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/internal/parkflag"

	goai "github.com/diagridio/go-ai"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/mcp"
)

// RunSingle builds a one-node agent around node, registers it in Catalyst, runs
// it once, and prints the result. tools may be nil for agents that call no MCP
// tools; when set they are declared for the Catalyst Agents view.
func RunSingle(ctx context.Context, name, role, goal string, node agent.NodeFunc, tools []goai.ToolInfo) error {
	graph := agent.NewGraph(name).
		AddNode("act", node).
		SetEntry("act").
		AddEdge("act", agent.END)

	runner, err := goai.NewRunner(ctx, goai.Config{
		Graph:     graph,
		Name:      name,
		Framework: "langchaingo", // label only; shapes the workflow name
		Role:      role,
		Goal:      goal,
		Tools:     tools,
	})
	if err != nil {
		return err
	}
	defer runner.Close()

	out, err := runner.Invoke(ctx, agent.State{})
	if err != nil {
		return err
	}
	var r struct {
		Output string `json:"output"`
	}
	_ = out.Into(&r)
	fmt.Printf("Agent %q ran as workflow %s: %s\n", runner.Name(), runner.WorkflowName(), parkflag.Redact(r.Output))
	return nil
}

// FindTool locates a discovered MCP tool by name.
func FindTool(tools []mcp.Tool, name string) (mcp.Tool, error) {
	for _, t := range tools {
		if t.Name == name {
			return t, nil
		}
	}
	return mcp.Tool{}, fmt.Errorf("MCP tool %q not found — is the park-systems MCP server registered and scoped to this agent?", name)
}

// ToolInfos declares discovered tools for the Catalyst Agents view.
func ToolInfos(tools []mcp.Tool) []goai.ToolInfo {
	out := make([]goai.ToolInfo, len(tools))
	for i, t := range tools {
		args, _ := json.Marshal(t.Schema)
		out[i] = goai.ToolInfo{Name: t.Name, Description: t.Description, Args: string(args)}
	}
	return out
}

// Fatal prints an error and exits non-zero.
func Fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
