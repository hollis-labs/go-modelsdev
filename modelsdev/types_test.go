package modelsdev_test

import (
	"encoding/json"
	"testing"

	"github.com/hollis-labs/go-modelsdev/modelsdev"
	"github.com/stretchr/testify/require"
)

// TestModel_UnmarshalAPIShape verifies that a Model decodes the on-the-wire
// models.dev shape — flat capability bools, "modalities" key, "knowledge"
// key, "context"/"output" Limit keys.
func TestModel_UnmarshalAPIShape(t *testing.T) {
	// Trimmed slice of the live https://models.dev/api.json response for
	// the anthropic/claude-sonnet-4-5 model, captured 2026-05-10.
	raw := []byte(`{
		"id": "claude-sonnet-4-5",
		"name": "Claude Sonnet 4.5",
		"family": "claude",
		"attachment": true,
		"reasoning": true,
		"tool_call": true,
		"temperature": true,
		"knowledge": "2025-07",
		"release_date": "2025-09-29",
		"last_updated": "2025-09-29",
		"modalities": {"input": ["text", "image"], "output": ["text"]},
		"open_weights": false,
		"cost": {"input": 3.0, "output": 15.0},
		"limit": {"context": 200000, "output": 64000}
	}`)

	var m modelsdev.Model
	require.NoError(t, json.Unmarshal(raw, &m))

	require.Equal(t, "claude-sonnet-4-5", m.ID)
	require.Equal(t, "Claude Sonnet 4.5", m.Name)
	require.Equal(t, "2025-07", m.KnowledgeCutoff)
	require.Equal(t, 200000, m.Limit.ContextWindow)
	require.Equal(t, 64000, m.Limit.MaxOutputTokens)
	require.Equal(t, []string{"text", "image"}, m.Modality.Input)
	require.Equal(t, []string{"text"}, m.Modality.Output)
	require.True(t, m.Capabilities.ToolCall)
	require.True(t, m.Capabilities.Reasoning)
	require.True(t, m.Capabilities.Attachment)
	require.True(t, m.Capabilities.Temperature)
	require.Equal(t, 3.0, m.Cost.Input)
	require.Equal(t, 15.0, m.Cost.Output)
}

// TestModel_RoundTrip verifies Marshal then Unmarshal preserves all fields.
// The disk cache relies on this round-trip — if Marshal emits a different
// shape than Unmarshal expects, cached catalogues would silently lose data.
func TestModel_RoundTrip(t *testing.T) {
	original := modelsdev.Model{
		ID:              "fast-1",
		Name:            "Fast Model 1",
		Family:          "fast",
		OpenWeights:     true,
		ReleaseDate:     "2025-01-01",
		KnowledgeCutoff: "2024-12",
		LastUpdated:     "2025-02-01",
		Cost:            modelsdev.Pricing{Input: 1.0, Output: 2.0, CacheRead: 0.1},
		Limit:           modelsdev.Limits{ContextWindow: 128000, MaxOutputTokens: 4096},
		Modality:        modelsdev.Modality{Input: []string{"text"}, Output: []string{"text"}},
		Capabilities: modelsdev.Capabilities{
			ToolCall:    true,
			Temperature: true,
		},
	}

	encoded, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded modelsdev.Model
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	require.Equal(t, original, decoded)
}

// TestProvider_EnvList verifies Provider.Env decodes as []string.
// The v0.1.0 type was string and would fail against any real provider entry.
func TestProvider_EnvList(t *testing.T) {
	raw := []byte(`{
		"id": "anthropic",
		"name": "Anthropic",
		"env": ["ANTHROPIC_API_KEY"],
		"models": {}
	}`)

	var p modelsdev.Provider
	require.NoError(t, json.Unmarshal(raw, &p))
	require.Equal(t, []string{"ANTHROPIC_API_KEY"}, p.Env)
}
