# Changelog

## v0.3.0 — 2026-09-30

### Added

- `Client.Run(ctx)` — the refresher loop on the calling goroutine. It blocks
  until ctx is cancelled and returns only once no refresh is in flight, so a
  caller can wait for the refresher before closing whatever its
  `WithOnRefresh` callback writes to. `StartRefresher` is now `go c.Run(ctx)`
  and is otherwise unchanged.

### Fixed

- `Refresh` returns `ctx.Err()` when ctx is cancelled after the fetch
  completes, instead of writing the cache and calling the `WithOnRefresh`
  callback for a result the caller has stopped wanting.

## v0.2.0 — 2026-05-10

Public-release prep. Aligns the on-the-wire schema with the live `models.dev/api.json`, adds runnable examples, and polishes the package surface for `pkg.go.dev`.

### Breaking

- `Provider.Env` is now `[]string` (was `string`). The live API encodes
  one or more environment-variable names per provider. Any pre-1.0
  consumer reading `Env` directly will need to switch to slice access.
- `Model` JSON encoding now matches the live API:
  - `knowledge_cutoff` JSON key → `knowledge` (Go field `KnowledgeCutoff` unchanged).
  - `Limits.ContextWindow` JSON key → `context` (was `context_window`).
  - `Limits.MaxOutputTokens` JSON key → `output` (was `max_output_tokens`).
  - `Modality` decodes from API key `modalities` (was `modality`).
  - `Capabilities` (`tool_call`, `reasoning`, `attachment`, `temperature`) is
    populated from flat top-level booleans on the model object, not a nested
    `capabilities` sub-object. The Go-side struct stays nested for
    ergonomics; a custom `MarshalJSON` / `UnmarshalJSON` handles the
    flat ↔ nested conversion so disk-cache round-trips remain byte-stable.

The v0.1.0 release decoded against a synthetic test-fixture shape, not the
real API; `Refresh` against `https://models.dev/api.json` would fail to
decode `env` and silently zero out capability and limit data. This release
fixes that.

### Added

- `examples/lookup` — runnable single-fetch demo that resolves a model and
  prints pricing + context window.
- `examples/refresher` — runnable background-refresher demo that uses
  `WithOnRefresh` to wait for the first successful fetch.
- `Provider.Doc`, `Provider.NPM`, `Provider.API` — optional fields the API
  populates for some providers.
- `modelsdev/doc.go` — top-level package documentation block for
  `pkg.go.dev` rendering.
- `.gitignore` covering Go build artefacts plus agent-tooling scratch
  paths so they cannot be committed accidentally.

### Tests

- `TestModel_UnmarshalAPIShape` — decodes a captured slice of the live API
  to lock the on-the-wire contract.
- `TestModel_RoundTrip` — guarantees Marshal → Unmarshal is lossless,
  protecting the disk-cache format.
- `TestProvider_EnvList` — pins `Env` as `[]string`.

### Notes

- Existing `v0.1.0` disk caches are not forward-compatible (the schema
  changed). The cache is best-effort by design — failed decodes simply
  trigger a fresh fetch — but consumers can pre-emptively delete
  `~/.cache/go-modelsdev/catalog.json` after upgrading to skip one
  failed-load round trip.

## v0.1.0

- `Client` with `Get`, `GetProvider`, `List`, `ListProviders`, `Refresh`, `StartRefresher`, and `LastFetchedAt`.
- Disk cache with atomic write (write-to-tmp + rename) and configurable TTL (default 24h). Stale cache survives any failed fetch.
- Exponential backoff with up to 3 retry attempts on `Refresh`.
- Background refresh loop via `StartRefresher` — stale-check on boot, schedules next tick at `lastFetchedAt + TTL`.
- Functional options: `WithCacheDir`, `WithCacheTTL`, `WithHTTPClient`, `WithURL`, `WithOnRefresh`.
- `WithOnRefresh(func(*Client))` — callback fired after every successful fetch; used by consumers to push catalog data into their own registries without polling.
