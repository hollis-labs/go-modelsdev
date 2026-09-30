package modelsdev_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hollis-labs/go-modelsdev/modelsdev"
	"github.com/stretchr/testify/require"
)

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// A fetch that completes after its context is cancelled must not reach the
// cache or the OnRefresh callback: the caller may already be tearing down
// whatever the callback writes to.
func TestRefresh_CancelledAfterFetchCommitsNothing(t *testing.T) {
	body, err := json.Marshal(minimalCatalog())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The response is fully buffered, so decode succeeds; ctx is cancelled
	// in between the fetch and the commit, which is the shutdown race.
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		cancel()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    r,
		}, nil
	})}

	var calls atomic.Int32
	dir := t.TempDir()
	c := modelsdev.New(
		modelsdev.WithURL("http://modelsdev.invalid/api.json"),
		modelsdev.WithHTTPClient(hc),
		modelsdev.WithCacheDir(dir),
		modelsdev.WithOnRefresh(func(*modelsdev.Client) { calls.Add(1) }),
	)

	err = c.Refresh(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, calls.Load(), "OnRefresh must not run for a cancelled refresh")
	require.True(t, c.LastFetchedAt().IsZero(), "a cancelled refresh must not update the in-memory catalog")
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries, "a cancelled refresh must not write the cache dir")
}

// Run must not return while a refresh is in flight, and once it has
// returned nothing more reaches the cache dir or the callback.
func TestRun_ReturnsAfterInFlightRefreshOnCancel(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done() // hold the fetch open until the client gives up
	}))
	defer srv.Close()

	var calls atomic.Int32
	dir := filepath.Join(t.TempDir(), "cache")
	c := modelsdev.New(
		modelsdev.WithURL(srv.URL),
		modelsdev.WithHTTPClient(srv.Client()),
		modelsdev.WithCacheDir(dir),
		modelsdev.WithOnRefresh(func(*modelsdev.Client) { calls.Add(1) }),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // runs before srv.Close, which waits for the held handler
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Run(ctx)
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not start a refresh for a missing cache")
	}
	select {
	case <-done:
		t.Fatal("Run returned while its context was still live")
	default:
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after its context was cancelled")
	}

	require.Zero(t, calls.Load())
	_, err := os.Stat(dir)
	require.True(t, os.IsNotExist(err), "no cache write may happen for a cancelled refresh")
}

// With a fresh cache Run makes no fetch and simply waits for cancellation.
func TestRun_FreshCacheWaitsForCancel(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(minimalCatalog())
	}))
	defer srv.Close()

	dir := t.TempDir()
	require.NoError(t, modelsdev.SaveCacheForTest(dir, modelsdev.Catalog{Providers: minimalCatalog()}))

	c := modelsdev.New(
		modelsdev.WithURL(srv.URL),
		modelsdev.WithHTTPClient(srv.Client()),
		modelsdev.WithCacheDir(dir),
		modelsdev.WithCacheTTL(time.Hour),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Run(ctx)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after its context was cancelled")
	}
	require.Zero(t, hits.Load(), "a fresh cache needs no fetch")
	m, ok := c.Get("acme", "fast-1")
	require.True(t, ok)
	require.Equal(t, "Fast Model 1", m.Name)
}
