// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command ctf02-durable is GopherCon CTF challenge 2 — "Hold Onto Your Butts".
//
// A durable multi-node agent (reroute -> breakers -> restore) that runs on the
// Dapr Workflow backend. Give it a fixed run id, kill it partway through, and
// re-run: Catalyst resumes from the last completed node instead of restarting.
// Only the final node calls the park-systems restore_power tool, so the flag
// appears only after the run survives the crash.
//
// This agent is deterministic — it needs no model at all. Set ACCESS_CODE to the
// flag you earned in challenge 1. See README.md.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	goai "github.com/diagridio/go-ai"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/mcp"

	daprclient "github.com/dapr/go-sdk/client"
)

const agentName = "power-grid"

type state struct {
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
	restore, err := findTool(tools, "restore_power")
	if err != nil {
		fatal(err)
	}

	accessCode := os.Getenv("ACCESS_CODE")
	if accessCode == "" {
		fatal(fmt.Errorf("set ACCESS_CODE to the flag you earned in challenge 1"))
	}

	// reroute -> breakers -> restore. The intermediate nodes checkpoint before the
	// flag-bearing restore node, so a crash mid-run resumes without re-charging.
	graph := agent.NewGraph(agentName).
		AddNode("reroute", stepNode("reroute", "rerouted auxiliary power")).
		AddNode("breakers", stepNode("breakers", "reset the tripped breakers")).
		AddNode("restore", func(ctx context.Context, _ agent.State) (agent.State, error) {
			res, err := restore.Call(ctx, map[string]any{"access_code": accessCode})
			if err != nil {
				return nil, fmt.Errorf("restore_power: %w", err)
			}
			return agent.State{"output": res}, nil
		}).
		SetEntry("reroute").
		AddEdge("reroute", "breakers").
		AddEdge("breakers", "restore").
		AddEdge("restore", agent.END)

	runner, err := goai.NewRunner(ctx, goai.Config{
		Graph:     graph,
		Name:      agentName,
		Framework: "langchaingo", // label only; shapes the workflow name
		Role:      "Grid Engineer",
		Goal:      "Restore the park power grid and survive a crash",
		Tools:     toolInfos(tools),
	})
	if err != nil {
		fatal(err)
	}
	defer runner.Close()

	out, err := runner.Invoke(ctx, agent.State{}, resumeOpts()...)
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

// stepNode is a deterministic checkpointed step. The pause gives you a window to
// Ctrl-C mid-run so you can prove the workflow resumes from the last node.
func stepNode(channel, msg string) agent.NodeFunc {
	return func(ctx context.Context, _ agent.State) (agent.State, error) {
		fmt.Printf("... %s\n", msg)
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return agent.State{channel: msg}, nil
	}
}

// findTool locates a discovered MCP tool by name.
func findTool(tools []mcp.Tool, name string) (mcp.Tool, error) {
	for _, t := range tools {
		if t.Name == name {
			return t, nil
		}
	}
	return mcp.Tool{}, fmt.Errorf("MCP tool %q not found — is the park-systems MCP server registered and scoped to this agent?", name)
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

// resumeOpts pins a fixed workflow instance ID when GOAI_RUN is set, so a re-run
// resumes an interrupted run instead of starting a new one.
func resumeOpts() []goai.InvokeOptions {
	if id := os.Getenv("GOAI_RUN"); id != "" {
		return []goai.InvokeOptions{{InstanceID: id}}
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
