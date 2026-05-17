package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- HTTP embed provider ---

type httpEmbedProvider struct {
	client *http.Client
	config HTTPEmbedConfig
}

func newHTTPEmbedProvider(cfg HTTPEmbedConfig) (*httpEmbedProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
	}
	return &httpEmbedProvider{client: client, config: cfg}, nil
}

func (p *httpEmbedProvider) Embed(text string) ([]float32, error) {
	vs, err := p.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	return vs[0], nil
}

func (p *httpEmbedProvider) EmbedBatch(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Build request payload — auto-detect Ollama vs generic OpenAI-compatible format.
	payload := p.buildPayload(texts)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, p.config.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http embed: status %d", resp.StatusCode)
	}

	return p.parseResponse(resp.Body)
}

// buildPayload constructs the request body based on endpoint patterns.
func (p *httpEmbedProvider) buildPayload(texts []string) []byte {
	url := p.config.URL

	if strings.Contains(url, "ollama") {
		// Ollama /api/embeddings format.
		model := p.config.Model
		if model == "" {
			model = "unknown"
		}
		type req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		body, _ := json.Marshal(req{Model: model, Input: texts})
		return body
	}

	// Generic OpenAI-compatible format.
	model := p.config.Model
	if model == "" {
		model = "unknown"
	}
	type req struct {
		Model string `json:"model"`
		Input any    `json:"input"`
	}
	// Accept both single string or array.
	var input any = texts
	if len(texts) == 1 {
		input = texts[0]
	}
	body, _ := json.Marshal(req{Model: model, Input: input})
	return body
}

// parseResponse extracts embedding vectors from the response.
func (p *httpEmbedProvider) parseResponse(body io.Reader) ([][]float32, error) {
	url := p.config.URL

	// Try Ollama format first.
	if strings.Contains(url, "ollama") {
		var ollamaResp struct {
			Embedding []float32 `json:"embedding"`
		}
		if err := json.NewDecoder(body).Decode(&ollamaResp); err == nil && len(ollamaResp.Embedding) > 0 {
			return [][]float32{ollamaResp.Embedding}, nil
		}
	}

	// Try OpenAI-compatible batch format.
	var openAIResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(body).Decode(&openAIResp); err == nil && len(openAIResp.Data) > 0 {
		results := make([][]float32, len(openAIResp.Data))
		for i, d := range openAIResp.Data {
			results[i] = d.Embedding
		}
		return results, nil
	}

	return nil, fmt.Errorf("http embed: could not parse response")
}

func (p *httpEmbedProvider) Dimensions() int {
	if p == nil {
		return 0
	}
	return p.config.Dimensions
}

func (p *httpEmbedProvider) Close() error {
	return nil // http.Client doesn't need explicit close
}
