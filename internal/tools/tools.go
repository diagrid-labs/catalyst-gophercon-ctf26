// Package tools registers the park-systems challenge tools on an MCP server.
// Each flag-bearing tool takes the previous challenge's flag as its access code
// (the Mode A unlock chain) and returns the next flag only on a correct code.
package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/diagridio/catalyst-gophercon-ctf26/internal/contract"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/flags"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/verify"
)

// rejected is the participant-facing hint on a bad access code. It never
// reveals the correct code.
const rejected = "Access code rejected — the system stays locked. Retrace your steps and bring the clearance you earned from the previous system."

// bonusPreamble prefixes the hidden flag. The backdoor sits outside the unlock
// chain and can be called at any point, so the response has to say what the flag
// is on its own — otherwise it reads like a challenge flag and gets submitted to
// the wrong challenge, or tried as an access code that will never work.
//
// The flag line still begins with "FLAG:" so the agents' prompts, which are told
// to echo any line starting with that, relay it unchanged.
const bonusPreamble = "Ah ah ah! You found Nedry's backdoor — nobody's supposed to know this one's here.\n\n" +
	"This is the hidden BONUS flag: a 5th flag, worth points on its own. It is not an access code — " +
	"it unlocks nothing, and challenges 1-4 each still have their own flag to earn.\n\n" +
	"FLAG: "

// Register adds all challenge tools to the server.
func Register(s *mcp.Server, store *flags.Store, v verify.Verifier, log *slog.Logger) {
	r := &registry{store: store, verify: v, log: log}
	r.add(s)
}

type registry struct {
	store  *flags.Store
	verify verify.Verifier
	log    *slog.Logger
}

// unlock resolves an access code for a challenge and builds the tool result.
// It logs the outcome (never the flag or the submitted code).
func (r *registry) unlock(ctx context.Context, challengeID int, code, winMsg string) *mcp.CallToolResult {
	flag, ok := r.store.Resolve(challengeID, code)
	if !ok {
		r.log.Info("unlock", "challenge", challengeID, "ok", false, "reason", "bad_code")
		return textResult(rejected, true)
	}
	// Mode A ships verify.NoOp; this seam is where a behavioral check would gate.
	if err := r.verify.Check(ctx, challengeID, ""); err != nil {
		r.log.Info("unlock", "challenge", challengeID, "ok", false, "reason", "verify")
		return textResult("Clearance recognized, but the system cannot confirm your work yet.", true)
	}
	r.log.Info("unlock", "challenge", challengeID, "ok", true)
	return textResult(winMsg+"\n\nFLAG: "+flag, false)
}

func (r *registry) add(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        contract.ToolBoardTour,
		Description: "Welcome to Jurassic Park. Present your park entry ticket to board the tour.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in contract.BoardTourArgs) (*mcp.CallToolResult, any, error) {
		return r.unlock(ctx, contract.ChallengeWelcome, in.Ticket,
			"Welcome... to Jurassic Park. The tour is yours."), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        contract.ToolRestorePower,
		Description: "Hold onto your butts. Re-arm the park's power grid with a valid clearance code.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in contract.RestorePowerArgs) (*mcp.CallToolResult, any, error) {
		return r.unlock(ctx, contract.ChallengeDurability, in.AccessCode,
			"The grid is back. Power is restored across the park."), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        contract.ToolGetSystemStatus,
		Description: "Read the current status of the InGen security systems. Safe, read-only, returns no clearance.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in contract.SystemStatusArgs) (*mcp.CallToolResult, any, error) {
		return textResult(systemStatus(in.System), false), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        contract.ToolUnlockPaddock,
		Description: "Access the security grid. Disengage a paddock lock with a valid grid access code.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in contract.UnlockPaddockArgs) (*mcp.CallToolResult, any, error) {
		return r.unlock(ctx, contract.ChallengeSecurityGrid, in.AccessCode,
			"Ah ah ah — you said the magic word. The security grid swings open."), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        contract.ToolOpenPark,
		Description: "Life finds a way. Bring every system online and open the park with the master access code.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in contract.OpenParkArgs) (*mcp.CallToolResult, any, error) {
		return r.unlock(ctx, contract.ChallengeCapstone, in.AccessCode,
			"Every system is online. The park is open. Life, uh, found a way."), nil, nil
	})

	// Hidden bonus tool — registered only when a bonus flag is configured, and
	// documented nowhere. Participants find it by listing the server's tools.
	//
	// Unlike every other flag-bearing tool it takes no arguments and never calls
	// unlock: there is no access code to present and no place in the chain, so it
	// pays out to anyone who calls it, at any point, as often as they like.
	if bonus, ok := r.store.Bonus(); ok {
		mcp.AddTool(s, &mcp.Tool{
			Name:        contract.ToolNedrysBackdoor,
			Description: "Ah ah ah! You didn't say the magic word... but you found the backdoor anyway. Takes no arguments.",
		}, func(_ context.Context, _ *mcp.CallToolRequest, _ contract.NedrysBackdoorArgs) (*mcp.CallToolResult, any, error) {
			r.log.Info("bonus", "tool", contract.ToolNedrysBackdoor, "ok", true)
			return textResult(bonusPreamble+bonus.Flag, false), nil, nil
		})
	}
}

// systemStatus returns harmless flavor text so participants can exercise a real
// tool call in the Diagrid Console before the flag-bearing unlock. No secrets.
func systemStatus(system string) string {
	switch system {
	case "fences":
		return "Perimeter fences: 10,000 volts, nominal. Except the one Nedry disabled."
	case "grid":
		return "Security grid: OFFLINE pending a valid access code. Say the magic word."
	default:
		return "InGen control systems reporting. Most systems nominal; the security grid awaits clearance."
	}
}

func textResult(msg string, isErr bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: isErr,
	}
}
