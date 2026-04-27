package search

import "testing"

func TestQueryClassifier_Classify(t *testing.T) {
	qc := &QueryClassifier{}

	tests := []struct {
		name     string
		query    string
		expected QueryType
	}{
		// Keyword-only cases
		{"empty", "", QueryTypeKeyword},
		{"whitespace only", "   ", QueryTypeKeyword},
		{"single short token", "go", QueryTypeKeyword},
		{"single technical token", "internal/search", QueryTypeKeyword},
		{"file path", "handlers/search.go", QueryTypeKeyword},
		{"code identifier", "NewEmbedder", QueryTypeKeyword},
		{"@mention", "@task-123", QueryTypeKeyword},
		{"quoted phrase", "\"AND\"", QueryTypeKeyword},
		{"FTS5 AND operator", "foo AND bar", QueryTypeKeyword},
		{"FTS5 OR operator", "foo OR bar", QueryTypeKeyword},
		{"FTS5 NOT operator", "NOT foo", QueryTypeKeyword},
		{"single long technical token", "someverylongidentifierexample", QueryTypeKeyword},

		// Keyword: "and" is an FTS5 operator so "foo and bar" goes to keyword
		{"foo and bar (has AND)", "foo and bar", QueryTypeKeyword},

		// Semantic cases — long natural language (≥7 words) or very high stopword ratio (>75%)
		{"long phrase with stopwords", "what is the best way to implement search functionality in the system", QueryTypeSemantic},
		{"long natural language", "can you explain the difference between semantic and keyword search and how they differ", QueryTypeSemantic},
		{"very long query", "what are the best practices for configuring and using the embedding model for optimal search performance", QueryTypeSemantic},
		{"2-word question high ratio", "what is that", QueryTypeSemantic},

		// Hybrid cases — 2-6 words with stopwords, NOT matching FTS5 operators
		// "the foo" uses "the" which is not an FTS5 operator
		{"2-word with stopword", "the foo", QueryTypeHybrid},
		{"2-word question no operator", "what is NewEmbedder", QueryTypeHybrid},
		// "how NewEmbedder works" - "how" is a stopword but not an FTS5 operator
		{"short phrase with stopword", "how NewEmbedder works", QueryTypeHybrid},
		// "foo bar" - no stopwords, all technical → keyword
		{"2 words no stopword", "foo bar", QueryTypeKeyword},
		// "foo bar implementation" - 3 words, no stopwords, all technical → keyword
		{"3 words all technical", "foo bar implementation", QueryTypeKeyword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := qc.Classify(tt.query)
			if got != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.query, got, tt.expected)
			}
		})
	}
}

func TestQueryClassifier_CountWords(t *testing.T) {
	tests := []struct {
		query    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"hello, world", 2},
		{"one two three four five", 5},
		{"@task-123 foo", 2},
	}

	for _, tt := range tests {
		got := CountWords(tt.query)
		if got != tt.expected {
			t.Errorf("CountWords(%q) = %d, want %d", tt.query, got, tt.expected)
		}
	}
}

func TestQueryType_String(t *testing.T) {
	tests := []struct {
		qt       QueryType
		expected string
	}{
		{QueryTypeKeyword, "keyword"},
		{QueryTypeSemantic, "semantic"},
		{QueryTypeHybrid, "hybrid"},
		{QueryType(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.qt.String()
		if got != tt.expected {
			t.Errorf("QueryType(%d).String() = %q, want %q", tt.qt, got, tt.expected)
		}
	}
}

func TestGetRoutingMetrics(t *testing.T) {
	// Reset metrics for clean test state
	routingMetrics.Lock()
	routingMetrics.Keyword = 0
	routingMetrics.Semantic = 0
	routingMetrics.Hybrid = 0
	routingMetrics.Fallback = 0
	routingMetrics.Unlock()

	metrics := GetRoutingMetrics()
	if metrics["keyword"] != 0 {
		t.Errorf("expected keyword=0, got %d", metrics["keyword"])
	}
	if metrics["semantic"] != 0 {
		t.Errorf("expected semantic=0, got %d", metrics["semantic"])
	}
}

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"world", "world"},
		{"  ", ""},
		{"\t\ntest\t\n", "test"},
	}

	for _, tt := range tests {
		got := NormalizeQuery(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeQuery(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}