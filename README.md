# go-modelsdev

`go-modelsdev` is a Go library for consuming the [models.dev](https://models.dev) API. It provides a cached client that returns LLM provider and model metadata — pricing, context limits, modalities, and capabilities — for use in services that need to look up model info at runtime.

## Status

Beta — the API surface is small and stable.

## Install

```bash
go get github.com/hollis-labs/go-modelsdev
```

For local development with a `replace` directive, add to your `go.mod`:

```
replace github.com/hollis-labs/go-modelsdev => ../go-modelsdev
```

## Quick Usage

### Create a client and look up a model

```go
package main

import (
    "context"
    "fmt"

    "github.com/hollis-labs/go-modelsdev/modelsdev"
)

func main() {
    c := modelsdev.New()

    if err := c.Refresh(context.Background()); err != nil {
        panic(err)
    }

    m, ok := c.Get("anthropic", "claude-sonnet-4-5")
    if !ok {
        fmt.Println("model not found")
        return
    }
    fmt.Printf("%s — input: $%.2f/M tokens\n", m.Name, m.Cost.Input)
}
```

### Start the background refresher

The refresher keeps the cache fresh without blocking your application startup:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c := modelsdev.New()
c.StartRefresher(ctx) // returns immediately; refreshes in background

// c.Get, c.List, etc. are safe to call immediately (reads from disk cache
// if the in-memory catalog is not yet populated).
```

### List all models across all providers

```go
for _, ref := range c.List() {
    fmt.Printf("%s/%s — context window: %d\n",
        ref.ProviderID, ref.ID, ref.Limit.ContextWindow)
}
```

## Cache Behavior

- Default cache directory: `~/.cache/go-modelsdev/` (via `os.UserCacheDir()`)
- Cache file: `catalog.json`
- Default TTL: 24 hours
- **Atomic write contract**: the cache is written to a `.tmp` file first, then renamed into place. If a fetch or write fails, the existing cache remains intact and the client continues serving stale-but-valid data.

## Options

| Option | Default | Description |
|---|---|---|
| `WithCacheDir(dir string)` | `~/.cache/go-modelsdev/` | Override the cache directory |
| `WithCacheTTL(d time.Duration)` | `24h` | How long a cached catalog is considered fresh |
| `WithHTTPClient(c *http.Client)` | 30s-timeout client | Replace the HTTP client (useful for tests) |
| `WithURL(url string)` | `https://models.dev/api.json` | Override the API endpoint |

## Dependencies

- `github.com/cenkalti/backoff/v5` — exponential backoff for fetch retries (3 attempts)
- `github.com/stretchr/testify` — test assertions (test only)

## Testing

```bash
go test ./...
```

Tests use `httptest.NewServer` and `t.TempDir()` — no network access or real API keys required.

## License

MIT License. See `LICENSE`.
