// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package parkflag keeps earned flags out of terminal output.
//
// Every challenge prints what its agent returned, and that reply carries the
// flag. Participants are meant to read flags from the run's workflow output in
// the Catalyst console, so the value is stripped on its way to the terminal —
// the durable run still holds it, and the console is where it is claimed.
package parkflag

import (
	"regexp"
)

// flagPrefix is the CTF flag namespace. Flags look like gc26{<uuid>}.
const flagPrefix = "gc26"

// notice stands in for a stripped flag value.
const notice = "[withheld — read this run's workflow output in the Catalyst console]"

// flagPattern matches a whole flag value, wrapper included. It deliberately
// matches the value anywhere rather than only on a "FLAG:" line, since a model
// asked to echo a tool response may reflow or paraphrase around it.
var flagPattern = regexp.MustCompile(regexp.QuoteMeta(flagPrefix) + `\{[^}]*\}`)

// Redact replaces every flag value in text with a pointer to the console.
// Surrounding text is left alone, so an agent's "the grid swings open" still
// confirms the run worked.
func Redact(text string) string {
	return flagPattern.ReplaceAllString(text, notice)
}
