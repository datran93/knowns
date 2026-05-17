package search

import (
	"fmt"

	"github.com/datran93/knowns/internal/models"
)

// EmbedProvider is the interface for embedding providers.
// It supports both local (ONNX) and remote (HTTP API) providers.
type EmbedProvider interface {
	// Embed returns the embedding for a single text.
	Embed(text string) ([]float32, error)
	// EmbedBatch returns embeddings for multiple texts.
	EmbedBatch(texts []string) ([][]float32, error)
	// Dimensions returns the embedding vector dimensionality.
	Dimensions() int
	// Close releases provider resources.
	Close() error
}

// ProviderType distinguishes local (onnx) from remote (http) providers.
type ProviderType int

const (
	ProviderTypeONNX ProviderType = iota
	ProviderTypeHTTP
)

func (p ProviderType) String() string {
	switch p {
	case ProviderTypeONNX:
		return "onnx"
	case ProviderTypeHTTP:
		return "http"
	default:
		return "unknown"
	}
}

// HTTPEmbedConfig describes a remote HTTP embedding endpoint.
type HTTPEmbedConfig struct {
	URL            string // e.g. "http://localhost:11434/api/embeddings" for Ollama
	Model          string // model name to send in request
	Dimensions     int
	MaxTokens      int
	TimeoutSeconds int
}

// Validate checks that the HTTP config has required fields.
func (c HTTPEmbedConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("http embed config: URL is required")
	}
	if c.Dimensions <= 0 {
		return fmt.Errorf("http embed config: Dimensions must be positive")
	}
	return nil
}

// AutoSelectProvider returns the best available embedder given the config.
// It tries ONNX first if modelName is non-empty and resolves to a known model,
// otherwise returns an HTTP provider if endpoint is configured.
func AutoSelectProvider(semantic models.SemanticSearchSettings) (EmbedProvider, ProviderType, error) {
	if semantic.Provider == "onnx" || (semantic.Provider == "" && semantic.Model != "") {
		if _, ok := EmbeddingModels[semantic.Model]; ok {
			p, err := newONNXEmbedProvider(semantic)
			return p, ProviderTypeONNX, err
		}
	}
	if semantic.HTTPEndpoint != "" {
		cfg := HTTPEmbedConfig{
			URL:            semantic.HTTPEndpoint,
			Model:          semantic.Model,
			Dimensions:     semantic.Dimensions,
			MaxTokens:      semantic.MaxTokens,
			TimeoutSeconds: 120,
		}
		if err := cfg.Validate(); err != nil {
			return nil, ProviderTypeHTTP, err
		}
		p, err := newHTTPEmbedProvider(cfg)
		return p, ProviderTypeHTTP, err
	}
	return nil, ProviderTypeHTTP, fmt.Errorf("no embedding provider available: set model or httpEndpoint")
}
