// Package modelsdev provides a client for the models.dev API.
// It fetches a catalog of LLM providers and models with pricing, limits,
// modality, and capability data. Results are cached on disk (default TTL: 24h).
package modelsdev

// Catalog is the top-level result returned by the models.dev API.
type Catalog struct {
	Providers map[string]Provider
}

// Provider describes a single LLM provider and its available models.
type Provider struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Env    string            `json:"env"`
	Models map[string]Model  `json:"models"`
}

// Model describes a single LLM model within a provider.
type Model struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Family          string       `json:"family"`
	OpenWeights     bool         `json:"open_weights"`
	ReleaseDate     string       `json:"release_date"`
	KnowledgeCutoff string       `json:"knowledge_cutoff"`
	LastUpdated     string       `json:"last_updated"`
	Cost            Pricing      `json:"cost"`
	Limit           Limits       `json:"limit"`
	Modality        Modality     `json:"modality"`
	Capabilities    Capabilities `json:"capabilities"`
}

// Pricing holds per-million-token cost figures in USD.
// Fields absent from the API response are zero.
type Pricing struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheWrite float64 `json:"cache_write"`
	CacheRead  float64 `json:"cache_read"`
	Reasoning  float64 `json:"reasoning"`
}

// Limits holds the token-count constraints for a model.
type Limits struct {
	ContextWindow  int `json:"context_window"`
	MaxOutputTokens int `json:"max_output_tokens"`
}

// Modality describes which input and output modalities the model supports.
type Modality struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

// Capabilities describes the optional features a model supports.
type Capabilities struct {
	ToolCall    bool `json:"tool_call"`
	Reasoning   bool `json:"reasoning"`
	Attachment  bool `json:"attachment"`
	Temperature bool `json:"temperature"`
}

// ModelRef is a flat view of a Model that also carries its parent ProviderID.
// It is convenient when iterating all models across all providers.
type ModelRef struct {
	ProviderID string

	ID              string
	Name            string
	Family          string
	OpenWeights     bool
	ReleaseDate     string
	KnowledgeCutoff string
	LastUpdated     string
	Cost            Pricing
	Limit           Limits
	Modality        Modality
	Capabilities    Capabilities
}
