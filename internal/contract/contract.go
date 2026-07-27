// Package contract is the single source of truth for the MCP tool surface:
// tool names, argument shapes, and challenge identifiers. Both the production
// server and the dev/test harness import it, so the wire contract is provably
// identical across them.
package contract

// Tool names exposed by the park-systems MCP server.
const (
	ToolBoardTour       = "board_tour"        // Challenge 1 — Welcome to Jurassic Park
	ToolRestorePower    = "restore_power"     // Challenge 2 — Hold Onto Your Butts
	ToolGetSystemStatus = "get_system_status" // Challenge 3 — read-only, harmless, returns no flag
	ToolUnlockPaddock   = "unlock_paddock"    // Challenge 3 — Access the Security Grid (flag-bearing)
	ToolOpenPark        = "open_park"         // Challenge 4 — Life Finds a Way
)

// Challenge identifiers. Each flag-bearing tool maps to exactly one.
const (
	ChallengeWelcome      = 1
	ChallengeDurability   = 2
	ChallengeSecurityGrid = 3
	ChallengeCapstone     = 4
)

// BoardTourArgs unlocks Challenge 1. Ticket is the bootstrap code from the handout.
type BoardTourArgs struct {
	Ticket string `json:"ticket" jsonschema:"your park entry ticket code from the challenge instructions"`
}

// RestorePowerArgs unlocks Challenge 2. AccessCode is the Challenge 1 flag.
type RestorePowerArgs struct {
	AccessCode string `json:"access_code" jsonschema:"clearance code from the previous system (the flag you just earned)"`
}

// SystemStatusArgs drives the read-only Challenge 3 status probe. It returns no flag.
type SystemStatusArgs struct {
	System string `json:"system,omitempty" jsonschema:"optional subsystem to query, e.g. 'fences' or 'grid'"`
}

// UnlockPaddockArgs unlocks Challenge 3. AccessCode is the Challenge 2 flag.
type UnlockPaddockArgs struct {
	AccessCode string `json:"access_code" jsonschema:"grid access code from the previous system (the flag you just earned)"`
	Paddock    string `json:"paddock,omitempty" jsonschema:"optional paddock identifier, e.g. 'raptor'"`
}

// OpenParkArgs unlocks Challenge 4. AccessCode is the Challenge 3 flag.
type OpenParkArgs struct {
	AccessCode string `json:"access_code" jsonschema:"master access code from the security grid (the flag you just earned)"`
}
