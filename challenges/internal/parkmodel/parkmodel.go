// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package parkmodel selects and builds the chat model the CTF agents run on.
//
// It defaults to a local Ollama server so that playing the CTF costs nothing,
// and switches to OpenAI when a key is available. Both providers go through
// langchaingo's OpenAI client: Ollama serves an OpenAI-compatible API, and
// langchaingo's native Ollama provider silently drops tool calls — which every
// model-driven challenge depends on.
package parkmodel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Providers reported by [Model.Provider] and selected by MODEL_PROVIDER.
const (
	ProviderOllama = "ollama"
	ProviderOpenAI = "openai"
)

const (
	defaultOllamaModel   = "qwen2.5:7b"
	defaultOllamaBaseURL = "http://localhost:11434/v1"

	// Ollama ignores the API key, but the OpenAI-compatible client requires a
	// non-empty one.
	ollamaAPIKey = "ollama"

	defaultOpenAIModel = "gpt-4o"

	preflightTimeout = 5 * time.Second
)

// Model is the chat model an agent runs on. It satisfies [llms.Model], and its
// String method describes the choice for logging at startup.
type Model struct {
	llms.Model

	// Provider is [ProviderOllama] or [ProviderOpenAI].
	Provider string
	// Name is the model name, e.g. "qwen2.5:7b" or "gpt-4o".
	Name string
}

// String describes the model as, for example,
// "ollama/qwen2.5:7b (local, free)".
func (m *Model) String() string {
	switch m.Provider {
	case ProviderOllama:
		return fmt.Sprintf("%s/%s (local, free)", m.Provider, m.Name)
	case ProviderOpenAI:
		return fmt.Sprintf("%s/%s (billed to your OpenAI account)", m.Provider, m.Name)
	default:
		return fmt.Sprintf("%s/%s", m.Provider, m.Name)
	}
}

// New builds the chat model the environment selects — see [Provider] for how
// that choice is made. Whichever provider wins, the model must support tool
// calling: every flag comes back from a tool result.
func New() (*Model, error) {
	provider := Provider()

	var (
		name string
		opts []openai.Option
	)
	switch provider {
	case ProviderOpenAI:
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, errors.New("MODEL_PROVIDER=openai but OPENAI_API_KEY is not set")
		}
		name = env("OPENAI_MODEL", defaultOpenAIModel)
		opts = append(opts, openai.WithModel(name), openai.WithToken(key))
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			opts = append(opts, openai.WithBaseURL(baseURL))
		}

	case ProviderOllama:
		baseURL := env("OLLAMA_BASE_URL", defaultOllamaBaseURL)
		name = env("OLLAMA_MODEL", defaultOllamaModel)
		if err := preflight(baseURL, name); err != nil {
			return nil, err
		}
		opts = append(opts,
			openai.WithModel(name),
			openai.WithBaseURL(baseURL),
			openai.WithToken(ollamaAPIKey),
		)

	default:
		return nil, fmt.Errorf("unknown MODEL_PROVIDER %q: use %q or %q", provider, ProviderOllama, ProviderOpenAI)
	}

	m, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("build %s model: %w", provider, err)
	}
	return &Model{Model: m, Provider: provider, Name: name}, nil
}

// Provider reports which provider the environment selects. An explicit
// MODEL_PROVIDER wins; otherwise OpenAI is chosen only when a key is already
// exported. Set MODEL_PROVIDER=ollama to stay on the free local model even with
// OPENAI_API_KEY in your environment.
func Provider() string {
	if p := os.Getenv("MODEL_PROVIDER"); p != "" {
		return strings.ToLower(p)
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return ProviderOpenAI
	}
	return ProviderOllama
}

// preflight fails fast on the two things that usually go wrong on a first run:
// Ollama isn't running, or the model hasn't been pulled yet. Anything it can't
// interpret is left for the model call itself to report.
func preflight(baseURL, model string) error {
	ctx, cancel := context.WithTimeout(context.Background(), preflightTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("model server request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("no model server reachable at %s: %w\n"+
			"  Install Ollama (https://ollama.com/download), then run:\n"+
			"    ollama serve\n"+
			"    ollama pull %s", baseURL, err, model)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil || len(list.Data) == 0 {
		return nil
	}
	for _, m := range list.Data {
		if m.ID == model {
			return nil
		}
	}
	return fmt.Errorf("model %q is not available at %s — pull it with: ollama pull %s", model, baseURL, model)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
