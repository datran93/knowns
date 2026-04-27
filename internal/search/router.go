package search

import (
	"strings"
	"sync"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
)

// RoutingMetrics tracks query type routing counts.
type RoutingMetrics struct {
	Keyword  int
	Semantic int
	Hybrid   int
	Fallback int
}

// routingMetrics is package-level metrics state.
var routingMetrics struct {
	sync.Mutex
	Keyword  int
	Semantic int
	Hybrid   int
	Fallback int
}

// GetRoutingMetrics returns a snapshot of routing metrics.
func GetRoutingMetrics() map[string]int {
	routingMetrics.Lock()
	defer routingMetrics.Unlock()
	return map[string]int{
		"keyword":  routingMetrics.Keyword,
		"semantic": routingMetrics.Semantic,
		"hybrid":   routingMetrics.Hybrid,
		"fallback": routingMetrics.Fallback,
	}
}

func incrementMetric(metric *int) {
	routingMetrics.Lock()
	defer routingMetrics.Unlock()
	*metric++
}

// SearchRouter routes queries to the appropriate backend based on query type.
type SearchRouter struct {
	classifier *QueryClassifier
	engine     *Engine
	embedder   *Embedder // may be nil if not initialized
	vecStore   VectorStore
}

// NewSearchRouter creates a router with the given components.
func NewSearchRouter(store *storage.Store, embedder *Embedder, vecStore VectorStore) *SearchRouter {
	return &SearchRouter{
		classifier: &QueryClassifier{},
		engine:     NewEngine(store, embedder, vecStore),
		embedder:   embedder,
		vecStore:   vecStore,
	}
}

// Route executes the search using the appropriate backend based on query classification.
func (r *SearchRouter) Route(query string, mode string) (*RouterSearchResult, error) {
	qType := r.classifier.Classify(query)

	if mode == "keyword" || qType == QueryTypeKeyword {
		incrementMetric(&routingMetrics.Keyword)
		opts := SearchOptions{Query: query, Mode: "keyword", Limit: 20}
		results, err := r.engine.keywordSearch(query, opts)
		if err != nil {
			return nil, err
		}
		return &RouterSearchResult{QueryType: QueryTypeKeyword, Results: results}, nil
	}

	if qType == QueryTypeSemantic {
		if r.embedder == nil || r.vecStore == nil {
			incrementMetric(&routingMetrics.Fallback)
			opts := SearchOptions{Query: query, Mode: "keyword", Limit: 20}
			results, err := r.engine.keywordSearch(query, opts)
			if err != nil {
				return nil, err
			}
			return &RouterSearchResult{QueryType: QueryTypeSemantic, Results: results}, nil
		}
		incrementMetric(&routingMetrics.Semantic)
		opts := SearchOptions{Query: query, Mode: "semantic", Limit: 20}
		results, err := r.engine.semanticSearch(query, opts)
		if err != nil {
			incrementMetric(&routingMetrics.Fallback)
			kwResults, kwErr := r.engine.keywordSearch(query, opts)
			if kwErr != nil {
				return nil, err
			}
			return &RouterSearchResult{QueryType: QueryTypeSemantic, Results: kwResults}, nil
		}
		return &RouterSearchResult{QueryType: QueryTypeSemantic, Results: results}, nil
	}

	// Hybrid: run both and merge.
	if r.embedder == nil || r.vecStore == nil {
		incrementMetric(&routingMetrics.Keyword)
		opts := SearchOptions{Query: query, Mode: "keyword", Limit: 20}
		results, err := r.engine.keywordSearch(query, opts)
		if err != nil {
			return nil, err
		}
		return &RouterSearchResult{QueryType: QueryTypeHybrid, Results: results}, nil
	}
	incrementMetric(&routingMetrics.Hybrid)

	opts := SearchOptions{Query: query, Mode: "hybrid", Limit: 20}
	kwResults, kwErr := r.engine.keywordSearch(query, opts)
	semResults, semErr := r.engine.semanticSearch(query, opts)

	if kwErr != nil && semErr != nil {
		return nil, kwErr
	}
	if semErr != nil {
		return &RouterSearchResult{QueryType: QueryTypeHybrid, Results: kwResults}, nil
	}
	if kwErr != nil {
		return &RouterSearchResult{QueryType: QueryTypeHybrid, Results: semResults}, nil
	}

	merged := mergeResults(kwResults, semResults, 20)
	return &RouterSearchResult{QueryType: QueryTypeHybrid, Results: merged}, nil
}

// RouteAndSearch executes the full search using the router and returns plain results.
func (r *SearchRouter) RouteAndSearch(query string, mode string) ([]models.SearchResult, error) {
	qType := r.classifier.Classify(query)
	return r.routeWithQueryType(query, mode, qType, SearchOptions{})
}

// RouteAndSearchWithOptions uses the router to search with full options.
func (r *SearchRouter) RouteAndSearchWithOptions(query string, opts SearchOptions) ([]models.SearchResult, error) {
	qType := r.classifier.Classify(query)
	return r.routeWithQueryType(query, opts.Mode, qType, opts)
}

func (r *SearchRouter) routeWithQueryType(query string, mode string, qType QueryType, opts SearchOptions) ([]models.SearchResult, error) {
	// Respect explicit mode overrides.
	if mode == "keyword" {
		incrementMetric(&routingMetrics.Keyword)
		opts.Mode = "keyword"
		return r.engine.keywordSearch(query, opts)
	}

	if mode == "semantic" {
		if r.embedder == nil || r.vecStore == nil {
			incrementMetric(&routingMetrics.Fallback)
			opts.Mode = "keyword"
			return r.engine.keywordSearch(query, opts)
		}
		incrementMetric(&routingMetrics.Semantic)
		opts.Mode = "semantic"
		results, err := r.engine.semanticSearch(query, opts)
		if err != nil {
			incrementMetric(&routingMetrics.Fallback)
			opts.Mode = "keyword"
			return r.engine.keywordSearch(query, opts)
		}
		return results, nil
	}

	// Hybrid mode or auto — use query classification.
	switch qType {
	case QueryTypeKeyword:
		incrementMetric(&routingMetrics.Keyword)
		opts.Mode = "keyword"
		return r.engine.keywordSearch(query, opts)

	case QueryTypeSemantic:
		if r.embedder == nil || r.vecStore == nil {
			incrementMetric(&routingMetrics.Fallback)
			opts.Mode = "keyword"
			return r.engine.keywordSearch(query, opts)
		}
		incrementMetric(&routingMetrics.Semantic)
		opts.Mode = "semantic"
		results, err := r.engine.semanticSearch(query, opts)
		if err != nil {
			incrementMetric(&routingMetrics.Fallback)
			opts.Mode = "keyword"
			return r.engine.keywordSearch(query, opts)
		}
		return results, nil

	case QueryTypeHybrid:
		if r.embedder == nil || r.vecStore == nil {
			incrementMetric(&routingMetrics.Keyword)
			opts.Mode = "keyword"
			return r.engine.keywordSearch(query, opts)
		}
		incrementMetric(&routingMetrics.Hybrid)

		kwOpts := opts
		kwOpts.Mode = "keyword"
		semOpts := opts
		semOpts.Mode = "semantic"

		kwResults, kwErr := r.engine.keywordSearch(query, kwOpts)
		semResults, semErr := r.engine.semanticSearch(query, semOpts)

		if kwErr != nil && semErr != nil {
			return nil, kwErr
		}
		if semErr != nil {
			return kwResults, nil
		}
		if kwErr != nil {
			return semResults, nil
		}
		return mergeResults(kwResults, semResults, opts.Limit), nil

	default:
		incrementMetric(&routingMetrics.Keyword)
		opts.Mode = "keyword"
		return r.engine.keywordSearch(query, opts)
	}
}

// ClassifyQuery is a convenience alias for the classifier.
func (r *SearchRouter) ClassifyQuery(query string) QueryType {
	return r.classifier.Classify(query)
}

// HybridScoreThreshold returns the score threshold for hybrid result inclusion.
func HybridScoreThreshold() float64 {
	return 0.1
}

// NormalizeQuery cleans a query string for classification.
func NormalizeQuery(query string) string {
	return strings.TrimSpace(query)
}

// RouterSearchResult wraps search results with routing metadata.
type RouterSearchResult struct {
	QueryType QueryType
	Results   []models.SearchResult
}