// Package main is a runnable example of go-modelsdev that fetches the
// catalog once and looks up a single model.
//
// Run from the repo root:
//
//	go run ./examples/lookup
//
// Expected output (subject to upstream catalog changes):
//
//	anthropic/claude-sonnet-4-5 — context: 200000 tokens
//	  input:  $3.00/M tokens
//	  output: $15.00/M tokens
//
// The example uses a temp directory for the on-disk cache so it can run
// repeatedly without colliding with a host install. Real applications
// usually let the client default to ~/.cache/go-modelsdev.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hollis-labs/go-modelsdev/modelsdev"
)

const (
	provider = "anthropic"
	model    = "claude-sonnet-4-5"
)

func main() {
	cacheDir, err := os.MkdirTemp("", "modelsdev-example-*")
	if err != nil {
		log.Fatalf("temp cache dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)

	c := modelsdev.New(modelsdev.WithCacheDir(cacheDir))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Refresh(ctx); err != nil {
		log.Fatalf("refresh: %v", err)
	}

	m, ok := c.Get(provider, model)
	if !ok {
		fmt.Printf("model %s/%s not found in catalog\n", provider, model)
		return
	}

	fmt.Printf("%s/%s — context: %d tokens\n", provider, m.ID, m.Limit.ContextWindow)
	fmt.Printf("  input:  $%.2f/M tokens\n", m.Cost.Input)
	fmt.Printf("  output: $%.2f/M tokens\n", m.Cost.Output)
	if m.Capabilities.ToolCall {
		fmt.Println("  tool_call: yes")
	}
}
