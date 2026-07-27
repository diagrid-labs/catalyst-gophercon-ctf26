// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command grid-ranger is one of the challenge 4 capstone agents. It secures the
// park perimeter and registers itself under Agents in Catalyst.
package main

import (
	"context"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/ctf04-capstone/internal/capstone"
	"github.com/diagridio/go-ai/agent"
)

func main() {
	ctx := context.Background()
	node := func(context.Context, agent.State) (agent.State, error) {
		return agent.State{"output": "Perimeter fences online at 10,000 volts."}, nil
	}
	if err := capstone.RunSingle(ctx, "grid-ranger", "Grid Ranger",
		"Secure the park perimeter", node, nil); err != nil {
		capstone.Fatal(err)
	}
}
