// Command park-systems is the GopherCon 2026 Diagrid CTF flag-delivery MCP
// server. It exposes themed challenge tools over Streamable HTTP; each
// flag-bearing tool returns its flag only when given the previous challenge's
// flag as an access code (the Mode A unlock chain). Real flags are loaded at
// boot from a JSON file referenced by the environment and are never compiled in.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/diagridio/catalyst-gophercon-ctf26/internal/flags"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/server"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/tools"
	"github.com/diagridio/catalyst-gophercon-ctf26/internal/verify"
)

const (
	version           = "0.1.0"
	defaultAddr       = ":8080"
	defaultFlagsFile  = "fixtures/flags.dev.json"
	defaultDevToken   = "dev-token"
	defaultRatePerMin = 30
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := loadConfig()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	store, err := flags.LoadFile(cfg.flagsFile)
	if err != nil {
		log.Error("load flags", "err", err)
		os.Exit(1)
	}

	srv := mcp.NewServer(&mcp.Implementation{Name: "park-systems", Version: version}, nil)
	tools.Register(srv, store, verify.NoOp{}, log)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return srv },
		// Stateless: no session affinity, so any Diagrid app can call in.
		// DisableLocalhostProtection: park-systems is a remote, bearer-authed
		// server reached via a tunnel/reverse proxy that dials loopback; the
		// SDK's DNS-rebinding guard (localhost listener + non-localhost Host)
		// would otherwise 403 every proxied request. Auth + rate limiting are
		// what actually gate this server.
		&mcp.StreamableHTTPOptions{Stateless: true, DisableLocalhostProtection: true},
	)

	handler := server.New(server.Options{
		BearerToken: cfg.bearer,
		RatePerMin:  cfg.ratePerMin,
		MCPHandler:  mcpHandler,
	})

	httpSrv := &http.Server{
		Addr:              cfg.addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("park-systems listening", "addr", cfg.addr, "profile", cfg.profile, "version", version)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("serve", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

type config struct {
	addr       string
	flagsFile  string
	bearer     string
	profile    string
	ratePerMin int
}

// loadConfig reads configuration from the environment. In the prod profile a
// bearer token is required; in dev a placeholder token is used so the server
// runs out of the box against the committed fixtures.
func loadConfig() (config, error) {
	profile := getenv("PARK_PROFILE", "dev")
	c := config{
		addr:       getenv("PARK_ADDR", defaultAddr),
		flagsFile:  getenv("PARK_FLAGS_FILE", defaultFlagsFile),
		bearer:     os.Getenv("PARK_BEARER_TOKEN"),
		profile:    profile,
		ratePerMin: defaultRatePerMin,
	}

	if v := os.Getenv("PARK_RATE_PER_MIN"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return config{}, fmt.Errorf("PARK_RATE_PER_MIN %q: %w", v, err)
		}
		c.ratePerMin = n
	}

	if c.bearer == "" {
		if profile == "prod" {
			return config{}, errors.New("PARK_BEARER_TOKEN is required in the prod profile")
		}
		c.bearer = defaultDevToken
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
