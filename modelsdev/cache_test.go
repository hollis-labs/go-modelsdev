package modelsdev_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hollis-labs/go-modelsdev/modelsdev"
	"github.com/stretchr/testify/require"
)

func sampleCatalog() modelsdev.Catalog {
	return modelsdev.Catalog{
		Providers: map[string]modelsdev.Provider{
			"example": {
				ID:   "example",
				Name: "Example Provider",
				Models: map[string]modelsdev.Model{
					"m1": {ID: "m1", Name: "Model One"},
				},
			},
		},
	}
}

func TestSaveCache_NoTmpOnSuccess(t *testing.T) {
	dir := t.TempDir()
	err := modelsdev.SaveCacheForTest(dir, sampleCatalog())
	require.NoError(t, err)

	tmp := filepath.Join(dir, "catalog.json.tmp")
	_, statErr := os.Stat(tmp)
	require.True(t, os.IsNotExist(statErr), ".tmp must not exist after successful save")

	main := filepath.Join(dir, "catalog.json")
	_, statErr = os.Stat(main)
	require.NoError(t, statErr, "catalog.json must exist after successful save")
}

func TestSaveCache_OldPreservedOnFailure(t *testing.T) {
	dir := t.TempDir()

	// Write a known-good initial cache.
	initial := sampleCatalog()
	require.NoError(t, modelsdev.SaveCacheForTest(dir, initial))

	// Corrupt the directory by making catalog.json.tmp a directory, which will
	// cause os.Create to fail, simulating a write failure mid-save.
	tmp := filepath.Join(dir, "catalog.json.tmp")
	require.NoError(t, os.Mkdir(tmp, 0o755))

	// Attempt to save a different catalog — this must fail.
	bad := modelsdev.Catalog{Providers: map[string]modelsdev.Provider{}}
	err := modelsdev.SaveCacheForTest(dir, bad)
	require.Error(t, err)

	// The original catalog.json must still be intact and readable.
	f, err := os.Open(filepath.Join(dir, "catalog.json"))
	require.NoError(t, err)
	defer f.Close()

	type cacheEntry struct {
		Data modelsdev.Catalog `json:"data"`
	}
	var entry cacheEntry
	require.NoError(t, json.NewDecoder(f).Decode(&entry))
	require.Contains(t, entry.Data.Providers, "example", "original cache must be preserved after write failure")
}
