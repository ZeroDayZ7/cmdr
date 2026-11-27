# Command reference

## Build and test

```bash
go build -o bin/cmdr.exe
go test ./...
go test -race ./...
go vet ./...
golangci-lint run ./...
```

## Gemini ASK

CMDR ASK uses Google's Gemini Interactions API by default:

```text
POST https://generativelanguage.googleapis.com/v1beta/interactions
```

The legacy `generateContent` API remains supported for migrated or explicitly configured setups.

### API key setup

Use the `GEMINI_API_KEY` environment variable or the `apiKey` field in `%USERPROFILE%\.cmdr\.config.json`:

```powershell
$env:GEMINI_API_KEY="YOUR_API_KEY"
```

```bash
export GEMINI_API_KEY="YOUR_API_KEY"
```

CMDR never prints the key.

### Configuration

Inspect configuration:

```bash
cmdr ask config
```

Change the persistent default model:

```bash
cmdr ask config --model gemini-4-hydrogen
```

The modern configuration structure is:

```json
{
  "gemini": {
    "baseUrl": "https://generativelanguage.googleapis.com/v1beta",
    "apiKey": "",
    "model": "gemini-4-hydrogen",
    "api": "interactions"
  }
}
```

Fresh installations default to `gemini-4-hydrogen`, a stable general-purpose model documented by Google for coding and everyday agentic tasks.

Existing `gemini.url` configurations are migrated automatically. When a legacy model can be extracted from the endpoint, CMDR preserves the API key and model and keeps `generate-content` as the configured API. If migration is not possible, CMDR reports a warning and uses the modern Interactions API defaults.

### Ask

Use positional prompt text or `-q`:

```bash
cmdr ask "Explain this Go code"
cmdr ask -q "Explain this Go code"
```

Add file contents:

```bash
cmdr ask -q "Review these files" -f main.go -f config.go
```

Override the model for one request without changing the saved default:

```bash
cmdr ask --model gemini-3.8-flash "Review this code"
```

### Model discovery

List models compatible with the configured API:

```bash
cmdr ask models
```

Refresh the 24-hour model cache:

```bash
cmdr ask models --refresh
```

Model discovery uses paginated `GET /v1beta/models` requests and filters models using `supportedGenerationMethods`. Embedding, image, video, music and other non-text-generation models are not presented as text-generation options.