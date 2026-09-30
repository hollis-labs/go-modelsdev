package modelsdev

import (
	"context"
	"time"
)

// StartRefresher starts a background goroutine that keeps the catalog fresh.
//
// On start: if the cache is stale or missing, Refresh is called immediately.
// After each successful refresh, the next refresh is scheduled at
// lastFetchedAt + TTL so the goroutine drifts with actual fetch times
// rather than accumulating clock skew.
//
// Stops cleanly when ctx is cancelled. Safe to call from main().
// StartRefresher gives the caller nothing to wait on; use [Client.Run] in
// your own goroutine when shutdown must know the refresher has exited — for
// example before closing a store that [WithOnRefresh] writes to.
func (c *Client) StartRefresher(ctx context.Context) {
	go c.Run(ctx)
}

// Run is the refresher loop [Client.StartRefresher] starts, run on the
// calling goroutine. It blocks until ctx is cancelled and returns only once
// no refresh is in flight, so the cache write and the [WithOnRefresh]
// callback cannot happen after Run returns.
func (c *Client) Run(ctx context.Context) {
	// Determine whether an immediate refresh is needed.
	needNow := true
	if entry, err := loadCache(c.cacheDir); err == nil && !isStale(entry, c.ttl) {
		// Cache exists and is fresh — load it into memory and wait until it expires.
		c.mu.Lock()
		if c.catalog == nil {
			c.catalog = &entry.Data
			c.lastFetched = entry.FetchedAt
		}
		c.mu.Unlock()
		needNow = false
	}

	if needNow {
		// Run first refresh inline so the catalog is populated before we enter
		// the wait loop. Ignore errors — the old cache (if any) is still usable.
		_ = c.Refresh(ctx)
	}

	for {
		next := c.nextRefreshIn()
		select {
		case <-ctx.Done():
			return
		case <-time.After(next):
			_ = c.Refresh(ctx)
		}
	}
}

// nextRefreshIn returns how long until the next scheduled refresh.
// If the cache has never been populated it returns the full TTL.
func (c *Client) nextRefreshIn() time.Duration {
	c.mu.RLock()
	fetched := c.lastFetched
	c.mu.RUnlock()

	if fetched.IsZero() {
		return c.ttl
	}
	remaining := time.Until(fetched.Add(c.ttl))
	if remaining <= 0 {
		return 0
	}
	return remaining
}
