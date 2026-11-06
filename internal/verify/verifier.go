// Package verify is a dormant seam for behavioral verification (Mode B in the
// build plan). Mode A ships the NoOp verifier: the unlock chain alone gates the
// flags. A future Verifier could call the Diagrid management API to confirm a
// workflow crashed and resumed, or that several agents were orchestrated,
// before a flag is released — without changing the tool handlers.
package verify

import "context"

// Verifier optionally confirms that the real per-challenge work was performed.
// A nil error means the check passed and the flag may be released.
type Verifier interface {
	Check(ctx context.Context, challengeID int, projectID string) error
}

// NoOp always passes. It is the Mode A verifier.
type NoOp struct{}

// Check always returns nil.
func (NoOp) Check(context.Context, int, string) error { return nil }
