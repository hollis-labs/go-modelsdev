# go-modelsdev

A cached Go client for the models.dev API: provider and model metadata —
pricing, context limits, modalities, capabilities — for services that look up
model info at runtime. It fetches, caches and serves that catalog; it does not
call models, choose between them, or interpret their capabilities.

## Start Here

- `README.md` covers lookup and refresher usage.
- `modelsdev/client.go` owns `New`, `Refresh`, `Get` and `List`.
- `modelsdev/cache.go` owns on-disk caching and the atomic write.
- `modelsdev/refresher.go` owns background refresh and backoff.
- `modelsdev/types.go` maps the upstream API shape.
- `examples/lookup` and `examples/refresher` are runnable.

## Commands

```bash
gofmt -l .
go vet ./...
go test -race -count=1 ./...
```

Tests serve fixtures through `httptest` and never reach models.dev, so a run
needs no network. There is no CI workflow or Makefile in this repo.

## Boundaries

The cache must survive a bad refresh. Writes are atomic, a failed save leaves
the previous cache intact, and a successful one leaves no temp file behind —
`TestRefresh_AtomicWrite`, `TestSaveCache_OldPreservedOnFailure` and
`TestSaveCache_NoTmpOnSuccess`. A refresh failure falls back to what is already
cached (`TestRefresh_Fallback`) rather than emptying it, because a service that
loses its model catalog mid-flight cannot route requests.

`modelsdev/types.go` mirrors an upstream API this repo does not control —
flat capability booleans, a `modalities` key, a `knowledge` date. Unmarshalling
is pinned against a trimmed real response (`TestModel_UnmarshalAPIShape`), so
treat that fixture as the record of what models.dev actually sends rather than
what would be convenient.

The tracked `go.work` (`use .`) is a single-module workspace and contradicts
the convention in these repos of leaving `go.work` uncommitted. It is harmless
today; do not build cross-module overrides on it.
