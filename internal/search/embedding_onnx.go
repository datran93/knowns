package search

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/datran93/knowns/internal/models"
)

type onnxEmbedProvider struct {
	embedder   *Embedder
	dimensions int
}

func newONNXEmbedProvider(semantic models.SemanticSearchSettings) (*onnxEmbedProvider, error) {
	modelCfg, ok := EmbeddingModels[semantic.Model]
	if !ok {
		return nil, fmt.Errorf("unknown embedding model %q", semantic.Model)
	}

	cacheDir := semantic.ModelDir
	if cacheDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cacheDir = filepath.Join(home, ".knowns", "models")
		}
	}

	embedder, err := NewEmbedder(EmbedderConfig{
		ModelDir:   cacheDir,
		ModelName:  semantic.Model,
		Dimensions: semantic.Dimensions,
		MaxTokens:  semantic.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("init onnx embedder: %w", err)
	}

	dims := semantic.Dimensions
	if dims <= 0 {
		dims = modelCfg.Dimensions
	}
	if d := embedder.Dimensions(); d > 0 {
		dims = d
	}

	return &onnxEmbedProvider{embedder: embedder, dimensions: dims}, nil
}

func (p *onnxEmbedProvider) Embed(text string) ([]float32, error) {
	return p.embedder.Embed(text)
}

func (p *onnxEmbedProvider) EmbedBatch(texts []string) ([][]float32, error) {
	return p.embedder.EmbedBatch(texts)
}

func (p *onnxEmbedProvider) Dimensions() int {
	return p.dimensions
}

func (p *onnxEmbedProvider) Close() error {
	if p.embedder != nil {
		p.embedder.Close()
	}
	return nil
}
