package modelsdev

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
)

const (
	defaultURL = "https://models.dev/api.json"
	defaultTTL = 24 * time.Hour
	maxRetries = 3
)

// Client fetches and caches the models.dev catalog.
// All fields are private; use New and the functional options to configure.
type Client struct {
	url        string
	cacheDir   string
	ttl        time.Duration
	httpClient *http.Client

	mu          sync.RWMutex
	catalog     *Catalog
	lastFetched time.Time
}

// Option configures a Client.
type Option func(*Client)

// WithCacheDir overrides the directory where catalog.json is stored.
func WithCacheDir(dir string) Option {
	return func(c *Client) { c.cacheDir = dir }
}

// WithCacheTTL sets how long a cached catalog is considered fresh (default: 24h).
func WithCacheTTL(d time.Duration) Option {
	return func(c *Client) { c.ttl = d }
}

// WithHTTPClient replaces the HTTP client used for fetches (useful in tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithURL overrides the API endpoint (useful in tests or for alternate sources).
func WithURL(url string) Option {
	return func(c *Client) { c.url = url }
}

// New returns a Client with default settings applied before any opts.
func New(opts ...Option) *Client {
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		cacheBase = os.TempDir()
	}

	c := &Client{
		url:        defaultURL,
		cacheDir:   filepath.Join(cacheBase, "go-modelsdev"),
		ttl:        defaultTTL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Get returns the Model identified by providerID and modelID.
// The catalog is loaded from disk on first call if it has not been loaded yet.
func (c *Client) Get(providerID, modelID string) (Model, bool) {
	cat := c.ensureCatalog()
	if cat == nil {
		return Model{}, false
	}
	p, ok := cat.Providers[providerID]
	if !ok {
		return Model{}, false
	}
	m, ok := p.Models[modelID]
	return m, ok
}

// GetProvider returns the Provider identified by providerID.
func (c *Client) GetProvider(providerID string) (Provider, bool) {
	cat := c.ensureCatalog()
	if cat == nil {
		return Provider{}, false
	}
	p, ok := cat.Providers[providerID]
	return p, ok
}

// List returns a flat slice of all models across all providers, sorted
// by provider ID then model ID for stable output.
func (c *Client) List() []ModelRef {
	cat := c.ensureCatalog()
	if cat == nil {
		return nil
	}

	providerIDs := make([]string, 0, len(cat.Providers))
	for pid := range cat.Providers {
		providerIDs = append(providerIDs, pid)
	}
	sort.Strings(providerIDs)

	var refs []ModelRef
	for _, pid := range providerIDs {
		p := cat.Providers[pid]
		modelIDs := make([]string, 0, len(p.Models))
		for mid := range p.Models {
			modelIDs = append(modelIDs, mid)
		}
		sort.Strings(modelIDs)

		for _, mid := range modelIDs {
			m := p.Models[mid]
			refs = append(refs, ModelRef{
				ProviderID:      pid,
				ID:              m.ID,
				Name:            m.Name,
				Family:          m.Family,
				OpenWeights:     m.OpenWeights,
				ReleaseDate:     m.ReleaseDate,
				KnowledgeCutoff: m.KnowledgeCutoff,
				LastUpdated:     m.LastUpdated,
				Cost:            m.Cost,
				Limit:           m.Limit,
				Modality:        m.Modality,
				Capabilities:    m.Capabilities,
			})
		}
	}
	return refs
}

// ListProviders returns all providers sorted by ID.
func (c *Client) ListProviders() []Provider {
	cat := c.ensureCatalog()
	if cat == nil {
		return nil
	}
	ids := make([]string, 0, len(cat.Providers))
	for id := range cat.Providers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Provider, 0, len(ids))
	for _, id := range ids {
		out = append(out, cat.Providers[id])
	}
	return out
}

// LastFetchedAt returns when the cache was last successfully written.
// Returns the zero time if the cache has never been populated.
func (c *Client) LastFetchedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastFetched
}

// Refresh fetches fresh data from the API and updates the on-disk cache.
// It retries up to maxRetries times with exponential backoff.
// If the fetch fails, the existing cache remains untouched.
func (c *Client) Refresh(ctx context.Context) error {
	op := func() (Catalog, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
		if err != nil {
			return Catalog{}, backoff.Permanent(fmt.Errorf("modelsdev: build request: %w", err))
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return Catalog{}, fmt.Errorf("modelsdev: fetch: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return Catalog{}, fmt.Errorf("modelsdev: unexpected status %d", resp.StatusCode)
		}

		// The API returns a flat map of providerID → Provider; we wrap it in Catalog.
		var raw map[string]Provider
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return Catalog{}, backoff.Permanent(fmt.Errorf("modelsdev: decode response: %w", err))
		}
		return Catalog{Providers: raw}, nil
	}

	cat, err := backoff.Retry(ctx, op,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(maxRetries),
	)
	if err != nil {
		return err
	}

	if err := saveCache(c.cacheDir, cat); err != nil {
		return err
	}

	c.mu.Lock()
	c.catalog = &cat
	c.lastFetched = time.Now()
	c.mu.Unlock()
	return nil
}

// ensureCatalog returns the in-memory catalog, loading from disk if needed.
// Returns nil if no cache exists and the load fails.
func (c *Client) ensureCatalog() *Catalog {
	c.mu.RLock()
	if c.catalog != nil {
		cat := c.catalog
		c.mu.RUnlock()
		return cat
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check after acquiring write lock.
	if c.catalog != nil {
		return c.catalog
	}

	entry, err := loadCache(c.cacheDir)
	if err != nil {
		return nil
	}
	c.catalog = &entry.Data
	c.lastFetched = entry.FetchedAt
	return c.catalog
}
