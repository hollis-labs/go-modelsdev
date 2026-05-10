// Package modelsdev provides a cached client for the models.dev API.
//
// The client fetches a catalog of LLM providers and models — including
// pricing, context limits, modalities, and capabilities — and serves
// lookups from an in-memory copy backed by an on-disk cache.
//
// # Quickstart
//
//	c := modelsdev.New()
//	if err := c.Refresh(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	if m, ok := c.Get("anthropic", "claude-sonnet-4-5"); ok {
//	    fmt.Printf("%s — input: $%.2f/M tokens\n", m.Name, m.Cost.Input)
//	}
//
// # Cache behaviour
//
// The disk cache lives at ~/.cache/go-modelsdev/catalog.json by default
// (override with [WithCacheDir]). Writes are atomic — a temp file is
// rendered first and renamed into place — so a failed fetch never
// corrupts the cache. The default TTL is 24h ([WithCacheTTL] overrides).
//
// # Background refresh
//
// [Client.StartRefresher] runs a background goroutine that fetches
// immediately if the cache is stale, then schedules subsequent fetches
// at lastFetchedAt + TTL. Cancel the supplied context to stop cleanly.
//
// See the examples/ directory for runnable usage.
package modelsdev
