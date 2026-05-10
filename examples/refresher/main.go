// Package main is a runnable example of the go-modelsdev background
// refresher. It starts the refresh loop, waits for the first successful
// fetch via the WithOnRefresh hook, prints a summary of the catalog,
// and exits cleanly.
//
// Run from the repo root:
//
//	go run ./examples/refresher
//
// Expected output (subject to upstream catalog changes):
//
//	catalog refreshed: 42 providers, 318 models
//	first 3 providers:
//	  anthropic — Anthropic
//	  amazon-bedrock — Amazon Bedrock
//	  azure — Azure OpenAI
//
// The example uses a temp directory for the cache so it does not collide
// with a host install. Real services usually let the client default to
// ~/.cache/go-modelsdev and rely on the refresher to keep it fresh
// across the service's lifetime.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hollis-labs/go-modelsdev/modelsdev"
)

func main() {
	cacheDir, err := os.MkdirTemp("", "modelsdev-example-*")
	if err != nil {
		log.Fatalf("temp cache dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Buffered so the hook never blocks the refresher goroutine.
	refreshed := make(chan struct{}, 1)

	c := modelsdev.New(
		modelsdev.WithCacheDir(cacheDir),
		modelsdev.WithOnRefresh(func(_ *modelsdev.Client) {
			select {
			case refreshed <- struct{}{}:
			default:
			}
		}),
	)

	c.StartRefresher(ctx)

	select {
	case <-refreshed:
	case <-ctx.Done():
		log.Fatalf("waiting for first refresh: %v", ctx.Err())
	}

	providers := c.ListProviders()
	models := c.List()
	fmt.Printf("catalog refreshed: %d providers, %d models\n", len(providers), len(models))

	cap := 3
	if len(providers) < cap {
		cap = len(providers)
	}
	fmt.Printf("first %d providers:\n", cap)
	for _, p := range providers[:cap] {
		fmt.Printf("  %s — %s\n", p.ID, p.Name)
	}

	fmt.Printf("last fetched at: %s\n", c.LastFetchedAt().Format(time.RFC3339))
}
