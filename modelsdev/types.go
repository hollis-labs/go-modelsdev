package modelsdev

import "encoding/json"

// Catalog is the top-level result returned by the models.dev API.
type Catalog struct {
	Providers map[string]Provider
}

// Provider describes a single LLM provider and its available models.
//
// Env carries one or more environment-variable names the provider expects
// callers to set (e.g. ["ANTHROPIC_API_KEY"]). The API can return an empty
// list for providers that don't require a key.
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Env    []string         `json:"env"`
	Doc    string           `json:"doc,omitempty"`
	NPM    string           `json:"npm,omitempty"`
	API    string           `json:"api,omitempty"`
	Models map[string]Model `json:"models"`
}

// Model describes a single LLM model within a provider.
//
// The models.dev API encodes capability flags (tool_call, reasoning,
// attachment, temperature) as top-level booleans on the model object;
// this type folds them into the nested Capabilities struct for ergonomic
// access in Go. Modality is decoded from the API field "modalities".
type Model struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Family          string       `json:"family"`
	OpenWeights     bool         `json:"open_weights"`
	ReleaseDate     string       `json:"release_date"`
	KnowledgeCutoff string       `json:"knowledge"`
	LastUpdated     string       `json:"last_updated"`
	Cost            Pricing      `json:"cost"`
	Limit           Limits       `json:"limit"`
	Modality        Modality     `json:"-"`
	Capabilities    Capabilities `json:"-"`
}

// modelJSON is the on-the-wire shape of a Model — flat capability bools and
// the API's "modalities" key. Used by Model's MarshalJSON / UnmarshalJSON so
// the Go-side struct can keep capability and modality data nested for
// ergonomic access.
type modelJSON struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Family          string   `json:"family"`
	OpenWeights     bool     `json:"open_weights"`
	ReleaseDate     string   `json:"release_date"`
	KnowledgeCutoff string   `json:"knowledge"`
	LastUpdated     string   `json:"last_updated"`
	Cost            Pricing  `json:"cost"`
	Limit           Limits   `json:"limit"`
	Modalities      Modality `json:"modalities"`

	ToolCall    bool `json:"tool_call"`
	Reasoning   bool `json:"reasoning"`
	Attachment  bool `json:"attachment"`
	Temperature bool `json:"temperature"`
}

// UnmarshalJSON decodes the flat models.dev shape into the nested struct.
func (m *Model) UnmarshalJSON(b []byte) error {
	var j modelJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*m = Model{
		ID:              j.ID,
		Name:            j.Name,
		Family:          j.Family,
		OpenWeights:     j.OpenWeights,
		ReleaseDate:     j.ReleaseDate,
		KnowledgeCutoff: j.KnowledgeCutoff,
		LastUpdated:     j.LastUpdated,
		Cost:            j.Cost,
		Limit:           j.Limit,
		Modality:        j.Modalities,
		Capabilities: Capabilities{
			ToolCall:    j.ToolCall,
			Reasoning:   j.Reasoning,
			Attachment:  j.Attachment,
			Temperature: j.Temperature,
		},
	}
	return nil
}

// MarshalJSON re-emits the flat models.dev shape so cache round-trips are
// byte-stable: the disk cache and the API response decode identically.
func (m Model) MarshalJSON() ([]byte, error) {
	j := modelJSON{
		ID:              m.ID,
		Name:            m.Name,
		Family:          m.Family,
		OpenWeights:     m.OpenWeights,
		ReleaseDate:     m.ReleaseDate,
		KnowledgeCutoff: m.KnowledgeCutoff,
		LastUpdated:     m.LastUpdated,
		Cost:            m.Cost,
		Limit:           m.Limit,
		Modalities:      m.Modality,
		ToolCall:        m.Capabilities.ToolCall,
		Reasoning:       m.Capabilities.Reasoning,
		Attachment:      m.Capabilities.Attachment,
		Temperature:     m.Capabilities.Temperature,
	}
	return json.Marshal(j)
}

// Pricing holds per-million-token cost figures in USD.
// Fields absent from the API response are zero.
type Pricing struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheWrite float64 `json:"cache_write,omitempty"`
	CacheRead  float64 `json:"cache_read,omitempty"`
	Reasoning  float64 `json:"reasoning,omitempty"`
}

// Limits holds the token-count constraints for a model.
type Limits struct {
	ContextWindow   int `json:"context"`
	MaxOutputTokens int `json:"output"`
}

// Modality describes which input and output modalities the model supports.
type Modality struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

// Capabilities describes the optional features a model supports. Decoded
// from flat top-level booleans on the API model object.
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
