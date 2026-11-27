package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func setGeminiTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestLoadConfigCreatesModernDefault(t *testing.T) {
	setGeminiTestHome(t)

	cfg, warnings, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if cfg.BaseURL != DefaultBaseURL || cfg.Model != DefaultModel || cfg.API != InteractionsAPI {
		t.Fatalf("unexpected default config: %#v", cfg)
	}
}

func TestLegacyConfigMigration(t *testing.T) {
	setGeminiTestHome(t)

	path, err := profiles.GetAskConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	legacy := []byte(`{"gemini":{"url":"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent","apiKey":"preserved-key"}}`)
	if err := os.WriteFile(path, legacy, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, warnings, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.APIKey != "preserved-key" || cfg.Model != "gemini-2.0-flash" || cfg.API != GenerateContentAPI {
		t.Fatalf("unexpected migrated config: %#v", cfg)
	}
	if len(warnings) == 0 {
		t.Fatal("expected migration warning")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("migrated config is invalid JSON")
	}
}

func TestSaveAndLoadPersistentModel(t *testing.T) {
	setGeminiTestHome(t)

	cfg := DefaultConfig()
	cfg.Model = "gemini-2.5-pro"
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, _, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "gemini-2.5-pro" {
		t.Fatalf("loaded model = %q", loaded.Model)
	}
}

func TestModelPaginationAndFiltering(t *testing.T) {
	setGeminiTestHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Goog-Api-Key") != "test-key" {
			t.Fatal("missing API key header")
		}
		if r.URL.Query().Get("pageToken") == "next" {
			if err := json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]any{"name": "models/gemini-2.5-pro", "displayName": "Gemini 2.5 Pro", "supportedGenerationMethods": []string{"generateContent"}}}}); err != nil {
				t.Fatalf("encode model page: %v", err)
			}
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]any{"name": "models/gemini-embedding-001", "supportedGenerationMethods": []string{"embedContent"}}, map[string]any{"name": "models/gemini-2.5-flash", "displayName": "Gemini 2.5 Flash", "supportedGenerationMethods": []string{"generateContent"}}}, "nextPageToken": "next"}); err != nil {
			t.Fatalf("encode model page: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	models, err := client.ListModels(context.Background(), cfg, true)
	if err != nil {
		t.Fatalf("ListModels() error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("filtered model count = %d, want 2", len(models))
	}
}

func TestInteractionsRequestAndResponse(t *testing.T) {
	setGeminiTestHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/interactions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-key" {
			t.Fatal("missing API key header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "gemini-2.5-flash" || body["input"] != "hello" {
			t.Fatalf("unexpected body: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"output_text": "Hello from Gemini"}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	text, err := NewClient().Generate(context.Background(), cfg, "gemini-2.5-flash", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello from Gemini" {
		t.Fatalf("response = %q", text)
	}
}

func TestGenerateContentRequestAndParse(t *testing.T) {
	payload, err := json.Marshal(map[string]any{"model": "gemini-2.5-pro", "input": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) == 0 {
		t.Fatal("empty payload")
	}

	text, err := ParseResponse([]byte(`{"candidates":[{"content":{"parts":[{"text":"hello"},{"text":" world"}]}}]}`), GenerateContentAPI)
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello world" {
		t.Fatalf("text = %q", text)
	}
}

func TestParseResponseRejectsInvalidAndEmptyResponses(t *testing.T) {
	cases := []struct {
		name string
		body string
		api  string
	}{
		{"invalid JSON", `{`, InteractionsAPI},
		{"missing interactions text", `{"outputs":[{"type":"model"}]}`, InteractionsAPI},
		{"empty interactions text", `{"output_text":""}`, InteractionsAPI},
		{"missing generate candidates", `{"candidates":[]}`, GenerateContentAPI},
		{"empty generate text", `{"candidates":[{"content":{"parts":[{"text":""}]}}]}`, GenerateContentAPI},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if text, err := ParseResponse([]byte(tc.body), tc.api); err == nil || strings.TrimSpace(text) != "" {
				t.Fatalf("text=%q err=%v", text, err)
			}
		})
	}
}

func TestAPIStatusErrors(t *testing.T) {
	unauthorized := apiStatusError(http.StatusUnauthorized, []byte(`{"error":{"message":"bad key"}}`))
	if unauthorized.Error() != "Gemini authentication failed (401): API key is missing or invalid" {
		t.Fatalf("unexpected 401 error: %v", unauthorized)
	}
	quota := apiStatusError(http.StatusTooManyRequests, nil)
	if quota.Error() == "" {
		t.Fatal("missing 429 error")
	}
}

func TestModelCacheIndependentOfWorkingDirectory(t *testing.T) {
	home := setGeminiTestHome(t)
	cacheDir, err := profiles.EnsureConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(cacheDir) != ".cmdr" {
		t.Fatalf("cache dir = %q", cacheDir)
	}
	if _, err := os.Stat(filepath.Join(home, ".cmdr", "models-cache.json")); !os.IsNotExist(err) {
		t.Fatalf("cache should not exist yet: %v", err)
	}
}

func TestGenerateRetries503ThenSucceeds(t *testing.T) {
	setGeminiTestHome(t)
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, `{"error":{"message":"temporarily unavailable"}}`, http.StatusServiceUnavailable)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"output_text": "recovered"}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.Sleep = func(context.Context, time.Duration) error { return nil }
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	text, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 2)
	if err != nil {
		t.Fatal(err)
	}
	if text != "recovered" || attempts != 2 {
		t.Fatalf("text=%q attempts=%d", text, attempts)
	}
}

func TestGenerate429QuotaNoRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		if err := json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "Quota exhausted", "details": []any{map[string]any{"reason": "QUOTA_EXHAUSTED"}}}}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.Sleep = func(context.Context, time.Duration) error { return nil }
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	_, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 3)
	if err == nil {
		t.Fatal("expected 429 error")
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d, want 1", attempts)
	}
}

func TestGenerate429WithRetryAfterRetries(t *testing.T) {
	attempts := 0
	var delays []time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			if err := json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "temporarily rate limited"}}); err != nil {
				t.Fatalf("encode response: %v", err)
			}
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"output_text": "ok"}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.Sleep = func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	_, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 2)
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || len(delays) != 1 || delays[0] != time.Second {
		t.Fatalf("attempts=%d delays=%v", attempts, delays)
	}
}

func TestGenerate401NoRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, `{"error":{"message":"invalid key"}}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "bad-key"
	_, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 3)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d, want 1", attempts)
	}
}

func TestGenerateMaximumRetryCount(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, `{"error":{"message":"temporarily unavailable"}}`, http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClient()
	client.Sleep = func(context.Context, time.Duration) error { return nil }
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	_, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 3)
	if err == nil {
		t.Fatal("expected final error")
	}
	if attempts != 4 {
		t.Fatalf("attempts=%d, want 4", attempts)
	}
}

func TestGenerateContextCanceledDuringBackoff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"temporarily unavailable"}}`, http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient()
	client.Sleep = func(ctx context.Context, delay time.Duration) error {
		cancel()
		return ctx.Err()
	}
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	_, err := client.Generate(ctx, cfg, "gemini-4-hydrogen", "hello", 2)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "canceled") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryClassifier(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		if !isRetryableStatus(status, "", nil) {
			t.Fatalf("status %d should be retryable", status)
		}
	}
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		if isRetryableStatus(status, "", nil) {
			t.Fatalf("status %d should not be retryable", status)
		}
	}
	if isRetryableStatus(http.StatusTooManyRequests, "", []byte(`{"error":{"details":[{"reason":"QUOTA_EXHAUSTED"}]}}`)) {
		t.Fatal("quota exhaustion should not be retried")
	}
	if !isRetryableStatus(http.StatusTooManyRequests, "1", nil) {
		t.Fatal("429 with Retry-After should be retried")
	}
}

func TestNetworkFailureRetryThenSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if err := json.NewEncoder(w).Encode(map[string]any{"output_text": "network recovered"}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test-key"
	client := NewClient()
	client.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("temporary network failure")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"output_text":"network recovered"}`)), Request: req}, nil
	})}
	client.Sleep = func(context.Context, time.Duration) error { return nil }

	text, err := client.Generate(context.Background(), cfg, "gemini-4-hydrogen", "hello", 1)
	if err != nil {
		t.Fatal(err)
	}
	if text != "network recovered" {
		t.Fatalf("text=%q", text)
	}
	server.Close()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
