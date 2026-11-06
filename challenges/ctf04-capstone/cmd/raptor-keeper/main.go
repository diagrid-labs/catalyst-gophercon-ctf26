// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Command raptor-keeper is one of the challenge 4 capstone agents. It brings the
// raptor paddock online and registers itself under Agents in Diagrid.
package main

import (
	"context"

	"github.com/diagridio/catalyst-gophercon-ctf26/challenges/ctf04-capstone/internal/capstone"
	"github.com/diagridio/go-ai/agent"
)

func main() {
	ctx := context.Background()
	node := func(context.Context, agent.State) (agent.State, error) {
		return agent.State{"output": "Raptor paddock secured. Clever girl."}, nil
	}
	if err := capstone.RunSingle(ctx, "raptor-keeper", "Raptor Keeper",
		"Keep the raptor paddock contained", node, nil); err != nil {
		capstone.Fatal(err)
	}
}
