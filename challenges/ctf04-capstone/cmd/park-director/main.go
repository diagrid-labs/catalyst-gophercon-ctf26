// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command park-director is the lead challenge 4 capstone agent. With the other
// agents online, it opens the park by calling the park-systems open_park MCP
// tool with your access code, and reports the final flag. It registers itself
// under Agents in Diagrid alongside the others.
//
// Set ACCESS_CODE to the flag you earned in challenge 3. See README.md.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/ctf04-capstone/internal/capstone"
	"github.com/diagridio/go-ai/agent"
	"github.com/diagridio/go-ai/mcp"

	daprclient "github.com/dapr/go-sdk/client"
)

func main() {
	ctx := context.Background()

	dc, err := daprclient.NewClient()
	if err != nil {
		capstone.Fatal(fmt.Errorf("dapr client: %w", err))
	}
	defer dc.Close()
	wf, err := daprclient.NewWorkflowClient()
	if err != nil {
		capstone.Fatal(fmt.Errorf("workflow client: %w", err))
	}

	tools, err := mcp.DiscoverTools(ctx, dc, wf)
	if err != nil {
		capstone.Fatal(fmt.Errorf("discover MCP tools: %w", err))
	}
	openPark, err := capstone.FindTool(tools, "open_park")
	if err != nil {
		capstone.Fatal(err)
	}

	accessCode := os.Getenv("ACCESS_CODE")
	if accessCode == "" {
		capstone.Fatal(fmt.Errorf("set ACCESS_CODE to the flag you earned in challenge 3"))
	}

	node := func(ctx context.Context, _ agent.State) (agent.State, error) {
		res, err := openPark.Call(ctx, map[string]any{"access_code": accessCode})
		if err != nil {
			return nil, fmt.Errorf("open_park: %w", err)
		}
		return agent.State{"output": res}, nil
	}

	if err := capstone.RunSingle(ctx, "park-director", "Park Director",
		"Bring every system online and open the park", node, capstone.ToolInfos(tools)); err != nil {
		capstone.Fatal(err)
	}
}
