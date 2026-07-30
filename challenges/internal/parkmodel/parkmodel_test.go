// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

package parkmodel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// clearEnv isolates a test from whatever the developer has exported, since
// provider selection is driven entirely by the environment.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"MODEL_PROVIDER", "OLLAMA_MODEL", "OLLAMA_BASE_URL",
		"OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_BASE_URL",
	} {
		t.Setenv(key, "")
	}
}

// fakeOllama serves the OpenAI-compatible /models listing that preflight reads.
func fakeOllama(t *testing.T, models ...string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/models" {
			http.NotFound(w, req)
			return
		}
		entries := make([]string, len(models))
		for i, m := range models {
			entries[i] = `{"id":"` + m + `","object":"model"}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[` + strings.Join(entries, ",") + `]}`))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestProvider(t *testing.T) {
	tests := []struct {
		name          string
		modelProvider string
		apiKey        string
		want          string
	}{
		{"nothing set defaults to the free local model", "", "", ProviderOllama},
		{"an exported key opts into openai", "", "sk-test", ProviderOpenAI},
		{"explicit ollama beats an exported key", "ollama", "sk-test", ProviderOllama},
		{"explicit openai without a key", "openai", "", ProviderOpenAI},
		{"provider is case insensitive", "OpenAI", "", ProviderOpenAI},
		{"unknown provider reaches the error path", "bedrock", "", "bedrock"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("MODEL_PROVIDER", tt.modelProvider)
			t.Setenv("OPENAI_API_KEY", tt.apiKey)

			if got := Provider(); got != tt.want {
				t.Errorf("Provider() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewOllama(t *testing.T) {
	clearEnv(t)
	baseURL := fakeOllama(t, defaultOllamaModel)
	t.Setenv("OLLAMA_BASE_URL", baseURL)

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if m.Provider != ProviderOllama {
		t.Errorf("Provider = %q, want %q", m.Provider, ProviderOllama)
	}
	if m.Name != defaultOllamaModel {
		t.Errorf("Name = %q, want %q", m.Name, defaultOllamaModel)
	}
	if m.Model == nil {
		t.Error("embedded llms.Model is nil")
	}
}

func TestNewOllamaModelNotPulled(t *testing.T) {
	clearEnv(t)
	t.Setenv("OLLAMA_BASE_URL", fakeOllama(t, "some-other-model:7b"))
	t.Setenv("OLLAMA_MODEL", "missing-model:7b")

	_, err := New()
	if err == nil {
		t.Fatal("New() succeeded, want an error naming the missing model")
	}
	if !strings.Contains(err.Error(), "ollama pull missing-model:7b") {
		t.Errorf("error %q does not tell the user how to pull the model", err)
	}
}

func TestNewOllamaServerDown(t *testing.T) {
	clearEnv(t)
	// Port 1 is reserved and never listening, so this always fails to connect.
	t.Setenv("OLLAMA_BASE_URL", "http://localhost:1/v1")

	_, err := New()
	if err == nil {
		t.Fatal("New() succeeded, want an error about an unreachable server")
	}
	if !strings.Contains(err.Error(), "ollama serve") {
		t.Errorf("error %q does not tell the user to start Ollama", err)
	}
}

func TestNewOpenAI(t *testing.T) {
	clearEnv(t)
	t.Setenv("OPENAI_API_KEY", "sk-test-not-used-offline")

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if m.Provider != ProviderOpenAI {
		t.Errorf("Provider = %q, want %q", m.Provider, ProviderOpenAI)
	}
	if m.Name != defaultOpenAIModel {
		t.Errorf("Name = %q, want %q", m.Name, defaultOpenAIModel)
	}
}

func TestNewOpenAIWithoutKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("MODEL_PROVIDER", "openai")

	if _, err := New(); err == nil {
		t.Fatal("New() succeeded without OPENAI_API_KEY, want an error")
	}
}

func TestNewUnknownProvider(t *testing.T) {
	clearEnv(t)
	t.Setenv("MODEL_PROVIDER", "bedrock")

	_, err := New()
	if err == nil {
		t.Fatal("New() succeeded with an unknown provider, want an error")
	}
	if !strings.Contains(err.Error(), ProviderOllama) || !strings.Contains(err.Error(), ProviderOpenAI) {
		t.Errorf("error %q does not list the valid providers", err)
	}
}

func TestModelString(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		want  string
	}{
		{"ollama", Model{Provider: ProviderOllama, Name: "qwen2.5:7b"}, "ollama/qwen2.5:7b (local, free)"},
		{"openai", Model{Provider: ProviderOpenAI, Name: "gpt-4o"}, "openai/gpt-4o (billed to your OpenAI account)"},
		{"unknown", Model{Provider: "bedrock", Name: "nova"}, "bedrock/nova"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
