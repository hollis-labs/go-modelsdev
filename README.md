# go-modelsdev

`go-modelsdev` is a Go library for consuming the [models.dev](https://models.dev) API. It provides a cached client that returns LLM provider and model metadata — pricing, context limits, modalities, and capabilities — for use in services that need to look up model info at runtime.

[![Go Reference](https://pkg.go.dev/badge/github.com/hollis-labs/go-modelsdev.svg)](https://pkg.go.dev/github.com/hollis-labs/go-modelsdev)

## Status

Pre-1.0 (v0.x.y). The exported API surface is small and intended to remain stable, but breaking changes can still land in minor bumps before v1.0 — see [`CHANGELOG.md`](CHANGELOG.md) for each release's contract.

## Install

```bash
go get github.com/hollis-labs/go-modelsdev
```

## Quick Usage

### Look up a single model

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/hollis-labs/go-modelsdev/modelsdev"
)

func main() {
    c := modelsdev.New()

    if err := c.Refresh(context.Background()); err != nil {
        log.Fatal(err)
    }

    m, ok := c.Get("anthropic", "claude-sonnet-4-5")
    if !ok {
        fmt.Println("model not found")
        return
    }
    fmt.Printf("%s — input: $%.2f/M tokens\n", m.Name, m.Cost.Input)
}
```

### Run the background refresher

The refresher keeps the cache fresh without blocking your application startup:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c := modelsdev.New()
c.StartRefresher(ctx) // returns immediately; refreshes in background

// c.Get, c.List, etc. are safe to call immediately (reads from disk cache
// if the in-memory catalog is not yet populated).
```

`StartRefresher` gives you nothing to wait on. When shutdown must know the refresher has stopped — before closing a store your `WithOnRefresh` callback writes to, say — run the loop yourself with `Run`, which blocks until ctx is cancelled and returns only once no refresh is in flight:

```go
done := make(chan struct{})
go func() {
    defer close(done)
    c.Run(ctx)
}()

// on shutdown
cancel()
<-done
```

Use `WithOnRefresh` to be notified when the catalog has been updated — useful for pushing data into your own registry without polling:

```go
c := modelsdev.New(modelsdev.WithOnRefresh(func(c *modelsdev.Client) {
    log.Printf("catalog refreshed: %d models", len(c.List()))
}))
c.StartRefresher(ctx)
```

### List all models across all providers

```go
for _, ref := range c.List() {
    fmt.Printf("%s/%s — context window: %d\n",
        ref.ProviderID, ref.ID, ref.Limit.ContextWindow)
}
```

### Runnable examples

See [`examples/`](examples/) — each example is its own `main` package that runs end-to-end against the live API:

- [`examples/lookup`](examples/lookup) — one-shot fetch + `Get` lookup.
- [`examples/refresher`](examples/refresher) — `StartRefresher` + `WithOnRefresh` hook, prints a catalog summary.

```bash
go run ./examples/lookup
go run ./examples/refresher
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
| `WithOnRefresh(fn func(*Client))` | nil | Callback fired after every successful refresh |

## Dependencies

- [`github.com/cenkalti/backoff/v5`](https://github.com/cenkalti/backoff) — exponential backoff for fetch retries (3 attempts).
- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) — test assertions (test only).

## Testing

```bash
go test -race ./...
```

Tests use `httptest.NewServer` and `t.TempDir()` — no network access or real API keys required.

## License

[MIT](LICENSE) © Hollis Labs.
