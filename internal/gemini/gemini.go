package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

const (
	DefaultBaseURL     = "https://generativelanguage.googleapis.com/v1beta"
	DefaultModel       = "gemini-4-hydrogen"
	InteractionsAPI    = "interactions"
	GenerateContentAPI = "generate-content"

	DefaultRetries = 2
	MaxRetries     = 5
	maxRetryDelay  = 8 * time.Second
)

type Config struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	API     string `json:"api"`
}

type apiConfigLegacy struct {
	URL    string `json:"url,omitempty"`
	APIKey string `json:"apiKey,omitempty"`
}

type serviceFile map[string]json.RawMessage

type Model struct {
	Name                       string   `json:"name"`
	BaseModelID                string   `json:"baseModelId,omitempty"`
	DisplayName                string   `json:"displayName,omitempty"`
	Description                string   `json:"description,omitempty"`
	InputTokenLimit            int64    `json:"inputTokenLimit,omitempty"`
	OutputTokenLimit           int64    `json:"outputTokenLimit,omitempty"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods,omitempty"`
}

func (m Model) ID() string {
	id := strings.TrimPrefix(m.Name, "models/")
	if id == "" {
		id = m.BaseModelID
	}
	return id
}

type modelCache struct {
	FetchedAt time.Time `json:"fetchedAt"`
	Models    []Model   `json:"models"`
}

type Client struct {
	HTTPClient *http.Client
	Sleep      func(context.Context, time.Duration) error
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Sleep:      sleepContext,
	}
}

func DefaultConfig() Config {
	return Config{BaseURL: DefaultBaseURL, APIKey: "", Model: DefaultModel, API: InteractionsAPI}
}

func NormalizeConfig(cfg Config) Config {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = DefaultModel
	}
	if strings.TrimSpace(cfg.API) == "" {
		cfg.API = InteractionsAPI
	}
	cfg.API = strings.ToLower(strings.TrimSpace(cfg.API))
	if cfg.API != GenerateContentAPI {
		cfg.API = InteractionsAPI
	}
	return cfg
}

func LoadConfig() (Config, []string, error) {
	path, err := profiles.GetAskConfigPath()
	if err != nil {
		return Config{}, nil, err
	}

	warnings := []string{}
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := SaveConfig(cfg); err != nil {
			return Config{}, nil, err
		}
		return cfg, warnings, nil
	}
	if err != nil {
		return Config{}, nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var services serviceFile
	if err := json.Unmarshal(content, &services); err != nil {
		return Config{}, nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	raw, exists := services["gemini"]
	if !exists {
		cfg := DefaultConfig()
		if err := SaveConfig(cfg); err != nil {
			return Config{}, nil, err
		}
		warnings = append(warnings, "No Gemini configuration found; created the default configuration.")
		return cfg, warnings, nil
	}

	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawFields); err != nil {
		return Config{}, nil, fmt.Errorf("failed to inspect Gemini configuration in %s: %w", path, err)
	}
	_, hasBaseURL := rawFields["baseUrl"]
	_, hasModel := rawFields["model"]
	_, hasAPI := rawFields["api"]
	if hasBaseURL || hasModel || hasAPI {
		var cfg Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return Config{}, nil, fmt.Errorf("failed to parse Gemini configuration in %s: %w", path, err)
		}
		cfg = NormalizeConfig(cfg)
		return cfg, warnings, nil
	}

	var cfg Config
	var legacy apiConfigLegacy
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return Config{}, nil, fmt.Errorf("failed to parse Gemini configuration in %s: %w", path, err)
	}

	cfg = DefaultConfig()
	cfg.APIKey = legacy.APIKey
	if strings.TrimSpace(legacy.URL) == "" {
		warnings = append(warnings, "Migrated legacy Gemini configuration to the modern structure while preserving the API key.")
	}
	if strings.TrimSpace(legacy.URL) != "" {
		if endpoint, model, ok := migrateLegacyURL(legacy.URL); ok {
			cfg.BaseURL = endpoint
			cfg.Model = model
			cfg.API = GenerateContentAPI
			warnings = append(warnings, "Migrated legacy gemini.url configuration to baseUrl/model/api while preserving generateContent and the API key.")
		} else {
			warnings = append(warnings, "Could not migrate legacy gemini.url automatically; using the default base URL and Interactions API. Preserve the existing API key.")
		}
	}

	if err := SaveConfig(cfg); err != nil {
		return Config{}, warnings, err
	}
	return cfg, warnings, nil
}

func migrateLegacyURL(rawURL string) (baseURL, model string, ok bool) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", "", false
	}
	re := regexp.MustCompile(`/models/([^/:]+)`)
	match := re.FindStringSubmatch(parsed.Path)
	if len(match) != 2 || strings.TrimSpace(match[1]) == "" {
		return "", "", false
	}
	model = match[1]
	baseURL = parsed.Scheme + "://" + parsed.Host + "/v1beta"
	return baseURL, model, true
}

func SaveConfig(cfg Config) error {
	cfg = NormalizeConfig(cfg)
	path, err := profiles.GetAskConfigPath()
	if err != nil {
		return err
	}

	services := serviceFile{}
	content, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(content, &services); err != nil {
			return fmt.Errorf("failed to parse existing config file %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing config file %s: %w", path, err)
	}

	encoded, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode Gemini configuration: %w", err)
	}
	services["gemini"] = encoded

	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config file %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}

func (c *Client) EffectiveAPIKey(cfg Config) (string, error) {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" || key == "your_gemini_api_key_here" {
		key = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	}
	if key == "" {
		return "", errors.New("gemini API key is missing; set apiKey in ~/.cmdr/.config.json or export GEMINI_API_KEY")
	}
	return key, nil
}

func (c *Client) ListModels(ctx context.Context, cfg Config, refresh bool) ([]Model, error) {
	key, err := c.EffectiveAPIKey(cfg)
	if err != nil {
		return nil, err
	}

	if !refresh {
		if models, err := readModelCache(); err == nil && len(models) > 0 {
			return FilterModels(models, cfg.API), nil
		}
	}

	models, err := c.fetchAllModels(ctx, cfg.BaseURL, key)
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, errors.New("gemini Models API returned an empty model list")
	}
	// Cache persistence is an optimization; model discovery remains usable if it fails.
	_ = writeModelCache(models)
	filtered := FilterModels(models, cfg.API)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no models support the configured %s API", cfg.API)
	}
	return filtered, nil
}

func readModelCache() ([]Model, error) {
	path, err := profiles.GetConfigDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(path, "models-cache.json")
	content, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	var cache modelCache
	if err := json.Unmarshal(content, &cache); err != nil {
		return nil, err
	}
	if time.Since(cache.FetchedAt) > 24*time.Hour {
		return nil, errors.New("model cache expired")
	}
	return cache.Models, nil
}

func writeModelCache(models []Model) error {
	dir, err := profiles.EnsureConfigDir()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(dir, "models-cache.json")
	cache := modelCache{FetchedAt: time.Now().UTC(), Models: models}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0644)
}

func (c *Client) fetchAllModels(ctx context.Context, baseURL, key string) ([]Model, error) {
	var all []Model
	var pageToken string
	for {
		endpoint := strings.TrimRight(baseURL, "/") + "/models"
		if pageToken != "" {
			u, err := url.Parse(endpoint)
			if err != nil {
				return nil, fmt.Errorf("invalid Models API base URL: %w", err)
			}
			q := u.Query()
			q.Set("pageToken", pageToken)
			u.RawQuery = q.Encode()
			endpoint = u.String()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Models API request: %w", err)
		}
		req.Header.Set("X-Goog-Api-Key", key)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, networkError(err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("close Models API response body: %w", err)
		}
		if readErr != nil {
			return nil, fmt.Errorf("failed to read Models API response: %w", readErr)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, apiStatusError(resp.StatusCode, body)
		}

		var page struct {
			Models        []Model `json:"models"`
			NextPageToken string  `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("invalid Models API response: %w", err)
		}
		all = append(all, page.Models...)
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}

	sort.Slice(all, func(i, j int) bool { return all[i].ID() < all[j].ID() })
	return all, nil
}

func FilterModels(models []Model, api string) []Model {
	filtered := make([]Model, 0, len(models))
	for _, model := range models {
		id := strings.ToLower(model.ID())
		if id == "" {
			continue
		}
		if strings.Contains(id, "embedding") || strings.Contains(id, "imagen") || strings.Contains(id, "veo") || strings.Contains(id, "lyria") || strings.Contains(id, "aqa") {
			continue
		}
		if api == GenerateContentAPI {
			if len(model.SupportedGenerationMethods) == 0 || !containsFold(model.SupportedGenerationMethods, "generateContent") {
				continue
			}
		} else {
			if len(model.SupportedGenerationMethods) > 0 && !containsFold(model.SupportedGenerationMethods, "generateContent") {
				continue
			}
			if !strings.HasPrefix(id, "gemini") && !strings.HasPrefix(id, "gemma") {
				continue
			}
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func containsFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(value, wanted) {
			return true
		}
	}
	return false
}

func (c *Client) Generate(ctx context.Context, cfg Config, model, prompt string, retryCounts ...int) (string, error) {
	key, err := c.EffectiveAPIKey(cfg)
	if err != nil {
		return "", err
	}
	retries := DefaultRetries
	if len(retryCounts) > 0 {
		retries = retryCounts[0]
	}
	if retries < 0 {
		retries = 0
	}
	if retries > MaxRetries {
		retries = MaxRetries
	}

	if strings.TrimSpace(model) == "" {
		model = cfg.Model
	}
	model = strings.TrimPrefix(strings.TrimSpace(model), "models/")

	var endpoint string
	var payload []byte
	if cfg.API == GenerateContentAPI {
		endpoint = strings.TrimRight(cfg.BaseURL, "/") + "/models/" + url.PathEscape(model) + ":generateContent"
		body := map[string]any{"contents": []map[string]any{{"parts": []map[string]string{{"text": prompt}}}}}
		payload, err = json.Marshal(body)
	} else {
		endpoint = strings.TrimRight(cfg.BaseURL, "/") + "/interactions"
		payload, err = json.Marshal(map[string]any{"model": model, "input": prompt, "store": false})
	}
	if err != nil {
		return "", fmt.Errorf("failed to encode Gemini request: %w", err)
	}

	for attempt := 0; attempt <= retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", contextError(err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("failed to create Gemini request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Goog-Api-Key", key)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if attempt < retries && isTemporaryNetworkError(err) {
				if err := c.waitForRetry(ctx, attempt, ""); err != nil {
					return "", err
				}
				continue
			}
			return "", networkError(err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		if closeErr := resp.Body.Close(); closeErr != nil && readErr == nil {
			readErr = closeErr
		}
		if readErr != nil {
			if attempt < retries && ctx.Err() == nil {
				if err := c.waitForRetry(ctx, attempt, ""); err != nil {
					return "", err
				}
				continue
			}
			return "", fmt.Errorf("failed to read Gemini response: %w", readErr)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if attempt < retries && isRetryableStatus(resp.StatusCode, resp.Header.Get("Retry-After"), body) {
				if err := c.waitForRetry(ctx, attempt, resp.Header.Get("Retry-After")); err != nil {
					return "", err
				}
				continue
			}
			return "", apiStatusError(resp.StatusCode, body)
		}

		text, err := ParseResponse(body, cfg.API)
		if err != nil {
			return "", err
		}
		return text, nil
	}

	return "", errors.New("gemini request failed after the maximum retry count")
}

func ParseResponse(body []byte, api string) (string, error) {
	if api == GenerateContentAPI {
		var response struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return "", fmt.Errorf("invalid Gemini generateContent response: %w", err)
		}
		var text strings.Builder
		if len(response.Candidates) > 0 {
			for _, part := range response.Candidates[0].Content.Parts {
				text.WriteString(part.Text)
			}
		}
		if strings.TrimSpace(text.String()) == "" {
			return "", errors.New("gemini generateContent returned no text")
		}
		return text.String(), nil
	}

	var response struct {
		OutputText string `json:"output_text"`
		Outputs    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("invalid Gemini Interactions API response: %w", err)
	}
	if strings.TrimSpace(response.OutputText) != "" {
		return response.OutputText, nil
	}
	var text strings.Builder
	for _, output := range response.Outputs {
		if strings.TrimSpace(output.Text) != "" && (output.Type == "" || strings.Contains(output.Type, "text")) {
			text.WriteString(output.Text)
		}
	}
	if strings.TrimSpace(text.String()) == "" {
		return "", errors.New("gemini Interactions API returned no text")
	}
	return text.String(), nil
}

func networkError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("gemini request timed out; check the network and try again")
	}
	return fmt.Errorf("cannot reach Gemini API: %w", err)
}

func apiStatusError(status int, body []byte) error {
	message := ""
	var googleError struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &googleError) == nil {
		message = strings.TrimSpace(googleError.Error.Message)
	}
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if len(message) > 500 {
		message = message[:500]
	}

	switch status {
	case http.StatusUnauthorized:
		return errors.New("gemini authentication failed (401): API key is missing or invalid")
	case http.StatusForbidden:
		if message != "" {
			return fmt.Errorf("gemini authentication/authorization failed (403): %s", message)
		}
		return errors.New("gemini authentication/authorization failed (403): API key is invalid or lacks access")
	case http.StatusNotFound:
		return errors.New("gemini model or API endpoint was not found (404); verify the model and API configuration")
	case http.StatusTooManyRequests:
		return errors.New("gemini rate limit/quota exhausted (429); wait and retry or inspect your Google AI quota")
	}
	if message != "" {
		return fmt.Errorf("gemini API request failed (%d): %s", status, message)
	}
	return fmt.Errorf("gemini API request failed (%d)", status)
}

func (c *Client) waitForRetry(ctx context.Context, attempt int, retryAfter string) error {
	delay := retryDelay(attempt, retryAfter)
	sleep := c.Sleep
	if sleep == nil {
		sleep = sleepContext
	}
	if err := sleep(ctx, delay); err != nil {
		return contextError(err)
	}
	return nil
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds >= 0 {
			delay := time.Duration(seconds) * time.Second
			if delay > maxRetryDelay {
				return maxRetryDelay
			}
			return delay
		}
		if date, err := http.ParseTime(retryAfter); err == nil {
			delay := time.Until(date)
			if delay < 0 {
				delay = 0
			}
			if delay > maxRetryDelay {
				return maxRetryDelay
			}
			return delay
		}
	}

	delay := time.Duration(1<<min(attempt, 4)) * 250 * time.Millisecond
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func contextError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("gemini request context deadline exceeded")
	}
	if errors.Is(err, context.Canceled) {
		return errors.New("gemini request canceled")
	}
	return err
}

func isTemporaryNetworkError(err error) bool {
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func isRetryableStatus(status int, retryAfter string, body []byte) bool {
	if status >= 500 && status <= 599 {
		return true
	}
	if status != http.StatusTooManyRequests {
		return false
	}

	var googleError struct {
		Error struct {
			Details []struct {
				Reason string `json:"reason"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &googleError) != nil {
		return retryAfter != ""
	}
	for _, detail := range googleError.Error.Details {
		reason := strings.ToUpper(detail.Reason)
		if strings.Contains(reason, "QUOTA_EXHAUSTED") {
			return false
		}
		if strings.Contains(reason, "RESOURCE_EXHAUSTED") {
			return retryAfter != ""
		}
	}
	message := strings.ToLower(string(body))
	if strings.Contains(message, "quota exceeded") || strings.Contains(message, "quota_exhausted") || strings.Contains(message, "daily limit") {
		return false
	}
	return true
}
