# Changelog

## v0.1.0

- `Client` with `Get`, `GetProvider`, `List`, `ListProviders`, `Refresh`, `StartRefresher`, and `LastFetchedAt`.
- Disk cache with atomic write (write-to-tmp + rename) and configurable TTL (default 24h). Stale cache survives any failed fetch.
- Exponential backoff with up to 3 retry attempts on `Refresh`.
- Background refresh loop via `StartRefresher` — stale-check on boot, schedules next tick at `lastFetchedAt + TTL`.
- Functional options: `WithCacheDir`, `WithCacheTTL`, `WithHTTPClient`, `WithURL`, `WithOnRefresh`.
- `WithOnRefresh(func(*Client))` — callback fired after every successful fetch; used by consumers to push catalog data into their own registries without polling.
