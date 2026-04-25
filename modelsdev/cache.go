package modelsdev

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheFile = "catalog.json"
	cacheTmp  = "catalog.json.tmp"
)

type cacheEntry struct {
	FetchedAt time.Time `json:"fetched_at"`
	Data      Catalog   `json:"data"`
}

func loadCache(dir string) (*cacheEntry, error) {
	f, err := os.Open(filepath.Join(dir, cacheFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entry cacheEntry
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		return nil, fmt.Errorf("modelsdev: decode cache: %w", err)
	}
	return &entry, nil
}

// saveCache writes data to disk atomically: it writes to a .tmp file first,
// then renames it into place. If anything fails, the existing cache file
// remains untouched.
func saveCache(dir string, data Catalog) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("modelsdev: create cache dir: %w", err)
	}

	tmp := filepath.Join(dir, cacheTmp)
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("modelsdev: create tmp cache: %w", err)
	}

	entry := cacheEntry{FetchedAt: time.Now(), Data: data}
	if encErr := json.NewEncoder(f).Encode(entry); encErr != nil {
		f.Close()
		os.Remove(tmp) // best-effort cleanup; old cache is still intact
		return fmt.Errorf("modelsdev: encode cache: %w", encErr)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("modelsdev: close tmp cache: %w", err)
	}

	dest := filepath.Join(dir, cacheFile)
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("modelsdev: rename cache: %w", err)
	}
	return nil
}

func isStale(entry *cacheEntry, ttl time.Duration) bool {
	return time.Since(entry.FetchedAt) > ttl
}
