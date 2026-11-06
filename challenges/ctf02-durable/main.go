// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command ctf02-durable is GopherCon CTF challenge 2 — "Hold Onto Your Butts".
//
// A durable multi-node agent (reroute -> breakers -> restore) that runs on the
// Dapr Workflow backend. Set FAIL_AT to trip a node on purpose: the backend
// retries just that node, replaying the completed ones from history rather than
// re-running them. Only the final node calls the park-systems restore_power
// tool, so the flag appears only after the run recovers.
//
// This agent is deterministic — it needs no model at all. Set ACCESS_CODE to the
// flag you earned in challenge 1. See README.md.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/internal/parkflag"

	goai "github.com/diagridio/go-ai"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/durable"
	"github.com/diagridio/go-ai/mcp"

	daprclient "github.com/dapr/go-sdk/client"
)

const agentName = "power-grid"

// failEnv names the node to fail on purpose, to show the retry recover it.
const failEnv = "FAIL_AT"

// tripped marks FAIL_AT as already fired, so the retry succeeds. Activities run
// on worker goroutines, so this is touched concurrently.
var tripped atomic.Bool

// nodeRetry keeps a failed run alive: the backend holds the wait between
// attempts, so a tripped node recovers without the run being restarted.
var nodeRetry = &durable.NodeRetryPolicy{
	MaxAttempts:        maxAttempts,
	InitialInterval:    retryInterval,
	BackoffCoefficient: 1, // flat, so every attempt waits the same
	MaxInterval:        retryInterval,
}

const (
	maxAttempts   = 20
	retryInterval = 5 * time.Second
)

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
	// flag-bearing restore node, so a failure part-way through costs only that node.
	graph := agent.NewGraph(agentName).
		AddNode("reroute", stepNode("reroute", "rerouted auxiliary power")).
		AddNode("breakers", stepNode("breakers", "reset the tripped breakers")).
		AddNode("restore", func(ctx context.Context, _ agent.State) (agent.State, error) {
			if err := failIfRequested("restore"); err != nil {
				return nil, err
			}
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
		NodeRetry: nodeRetry,
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
	fmt.Printf("Grid: %s\n", parkflag.Redact(result.Output))
}

// stepNode is a deterministic checkpointed step. Its "EXECUTING" line is the
// evidence to watch: a completed step must never print it twice, however many
// times a later step is retried.
func stepNode(channel, msg string) agent.NodeFunc {
	return func(ctx context.Context, _ agent.State) (agent.State, error) {
		if err := failIfRequested(channel); err != nil {
			return nil, err
		}
		fmt.Printf("... EXECUTING %s: %s\n", channel, msg)
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return agent.State{channel: msg}, nil
	}
}

// failIfRequested fails the node named by FAIL_AT on its first attempt only, so
// the retry recovers the run on its own.
//
// It must return an error, never os.Exit: a reported failure leaves the backend
// owning the retry, while a killed process strands the step in a run that can
// never finish.
func failIfRequested(node string) error {
	if os.Getenv(failEnv) != node {
		return nil
	}
	if tripped.Swap(true) {
		fmt.Printf("... retrying %s after the trip\n", node)
		return nil
	}
	fmt.Printf("\n=== %s=%s — tripping %q. Retry comes due in %s; sit tight. ===\n", failEnv, node, node, retryInterval)
	return fmt.Errorf("the grid tripped in %q (%s=%s)", node, failEnv, node)
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

// toolInfos declares the discovered tools for the Diagrid Agents view.
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
