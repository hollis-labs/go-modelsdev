# Changelog

## v0.1.0

- Initial release: `Client` with `Get`, `GetProvider`, `List`, `ListProviders`, `Refresh`, `StartRefresher`, and `LastFetchedAt`.
- Disk cache with atomic write (write-to-tmp + rename) and configurable TTL (default 24h).
- Exponential backoff with up to 3 retry attempts on `Refresh`.
- Background refresh loop via `StartRefresher` — schedules next tick at `lastFetchedAt + TTL`.
- Functional options: `WithCacheDir`, `WithCacheTTL`, `WithHTTPClient`, `WithURL`.
