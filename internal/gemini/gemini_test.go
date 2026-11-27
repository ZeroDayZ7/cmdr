package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
			json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]any{"name": "models/gemini-2.5-pro", "displayName": "Gemini 2.5 Pro", "supportedGenerationMethods": []string{"generateContent"}}}})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]any{"name": "models/gemini-embedding-001", "supportedGenerationMethods": []string{"embedContent"}}, map[string]any{"name": "models/gemini-2.5-flash", "displayName": "Gemini 2.5 Flash", "supportedGenerationMethods": []string{"generateContent"}}}, "nextPageToken": "next"})
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
		json.NewEncoder(w).Encode(map[string]any{"output_text": "Hello from Gemini"})
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
	cfg := DefaultConfig()
	cfg.API = GenerateContentAPI
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
