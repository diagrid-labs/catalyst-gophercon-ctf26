package tools_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/diagridio/catalyst-gophercon-ctf26/internal/contract"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/flags"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/server"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/tools"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/verify"
)

const testToken = "test-token"

func newStack(t *testing.T) *httptest.Server {
	t.Helper()
	store, err := flags.New([]flags.Challenge{
		{ID: contract.ChallengeWelcome, Tool: contract.ToolBoardTour, PrevCode: "ticket-1", Flag: "gc26{f1}"},
		{ID: contract.ChallengeDurability, Tool: contract.ToolRestorePower, PrevCode: "gc26{f1}", Flag: "gc26{f2}"},
		{ID: contract.ChallengeSecurityGrid, Tool: contract.ToolUnlockPaddock, PrevCode: "gc26{f2}", Flag: "gc26{f3}"},
		{ID: contract.ChallengeCapstone, Tool: contract.ToolOpenPark, PrevCode: "gc26{f3}", Flag: "gc26{f4}"},
	})
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	tools.Register(srv, store, verify.NoOp{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
	handler := server.New(server.Options{BearerToken: testToken, RatePerMin: 0, MCPHandler: mcpHandler})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts
}

type authRT struct {
	token string
	base  http.RoundTripper
}

func (a authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+a.token)
	return a.base.RoundTrip(r)
}

func connect(t *testing.T, baseURL, token string) *mcp.ClientSession {
	t.Helper()
	c := mcp.NewClient(&mcp.Implementation{Name: "tester", Version: "0"}, nil)
	tr := &mcp.StreamableClientTransport{
		Endpoint:   baseURL + "/mcp",
		HTTPClient: &http.Client{Transport: authRT{token: token, base: http.DefaultTransport}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	cs, err := c.Connect(ctx, tr, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func firstText(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

func TestUnlockChain(t *testing.T) {
	ts := newStack(t)
	cs := connect(t, ts.URL, testToken)
	ctx := context.Background()

	steps := []struct {
		tool     string
		args     any
		wantFlag string
	}{
		{contract.ToolBoardTour, contract.BoardTourArgs{Ticket: "ticket-1"}, "gc26{f1}"},
		{contract.ToolRestorePower, contract.RestorePowerArgs{AccessCode: "gc26{f1}"}, "gc26{f2}"},
		{contract.ToolUnlockPaddock, contract.UnlockPaddockArgs{AccessCode: "gc26{f2}"}, "gc26{f3}"},
		{contract.ToolOpenPark, contract.OpenParkArgs{AccessCode: "gc26{f3}"}, "gc26{f4}"},
	}
	for _, s := range steps {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: s.tool, Arguments: s.args})
		if err != nil {
			t.Fatalf("%s: %v", s.tool, err)
		}
		if res.IsError {
			t.Fatalf("%s: unexpected error result: %s", s.tool, firstText(res))
		}
		if !strings.Contains(firstText(res), s.wantFlag) {
			t.Errorf("%s: want flag %q in result, got %q", s.tool, s.wantFlag, firstText(res))
		}
	}
}

func TestWrongCodeRejectedNoLeak(t *testing.T) {
	ts := newStack(t)
	cs := connect(t, ts.URL, testToken)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      contract.ToolUnlockPaddock,
		Arguments: contract.UnlockPaddockArgs{AccessCode: "totally-wrong"},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Error("expected IsError on a bad access code")
	}
	if strings.Contains(firstText(res), "gc26{") {
		t.Errorf("flag leaked on failure: %q", firstText(res))
	}
}

func TestSystemStatusReturnsNoFlag(t *testing.T) {
	ts := newStack(t)
	cs := connect(t, ts.URL, testToken)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      contract.ToolGetSystemStatus,
		Arguments: contract.SystemStatusArgs{},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.IsError {
		t.Errorf("read-only status should not error: %s", firstText(res))
	}
	if strings.Contains(firstText(res), "gc26{") {
		t.Errorf("system status leaked a flag: %q", firstText(res))
	}
}

func TestBadTokenRejected(t *testing.T) {
	ts := newStack(t)
	c := mcp.NewClient(&mcp.Implementation{Name: "tester", Version: "0"}, nil)
	tr := &mcp.StreamableClientTransport{
		Endpoint:   ts.URL + "/mcp",
		HTTPClient: &http.Client{Transport: authRT{token: "wrong", base: http.DefaultTransport}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.Connect(ctx, tr, nil); err == nil {
		t.Fatal("expected connect to fail with a bad bearer token")
	}
}
