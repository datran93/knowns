package search

import (
	"regexp"
	"strings"
	"unicode"
)

// QueryType classifies the query for routing.
type QueryType int

const (
	QueryTypeKeyword QueryType = iota
	QueryTypeSemantic
	QueryTypeHybrid
)

func (qt QueryType) String() string {
	switch qt {
	case QueryTypeKeyword:
		return "keyword"
	case QueryTypeSemantic:
		return "semantic"
	case QueryTypeHybrid:
		return "hybrid"
	default:
		return "unknown"
	}
}

// QueryClassifier determines the best backend for a query using heuristics.
// It is stateless and safe for concurrent use.
type QueryClassifier struct{}

// stopwords are common English words that appear in natural language but not
// in targeted keyword searches.
var stopwords = []string{
	"the", "a", "an", "is", "are", "was", "were", "be", "been", "being",
	"have", "has", "had", "do", "does", "did", "will", "would", "could",
	"should", "may", "might", "must", "shall",
	"this", "that", "these", "those",
	"i", "you", "he", "she", "it", "we", "they",
	"what", "which", "who", "whom", "whose",
	"where", "when", "why", "how",
	"and", "or", "but", "if", "then", "else",
	"in", "on", "at", "to", "for", "of", "with", "by", "from", "as",
	"about", "into", "through", "during", "before", "after",
	"not", "no", "nor", "so", "very", "just", "only", "also",
}

// operatorPatterns matches FTS5 boolean operators that strongly indicate
// a keyword/structured query.
var operatorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bAND\b`),
	regexp.MustCompile(`(?i)\bOR\b`),
	regexp.MustCompile(`(?i)\bNOT\b`),
	regexp.MustCompile(`(?i)\bNEAR\b`),
	regexp.MustCompile(`(?i)^["*]|^["*].*["*]$`), // quoted phrase or glob at start/end
}

// technicalPatterns match identifiers that look like code/file/technical terms.
var technicalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_./\-:]*$`), // bare identifier (must start with letter)
	regexp.MustCompile(`\.(go|ts|js|py|md|json|yaml|yml|toml)$`), // file extensions
	regexp.MustCompile(`^[a-zA-Z0-9_]+::`), // qualified names
	regexp.MustCompile(`^[a-zA-Z0-9_]+\(`), // function call
	regexp.MustCompile(`^[@#][a-zA-Z0-9_]+`), // @mention or #tag (must start with @ or #)
}

// Classify returns the query type for routing decisions.
// Uses heuristic rules without expensive NLP.
func (qc *QueryClassifier) Classify(query string) QueryType {
	if query == "" {
		return QueryTypeKeyword
	}

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return QueryTypeKeyword
	}

	fields := strings.Fields(trimmed)

	// Rule 1: Very short single token without spaces → keyword.
	if len(fields) == 1 && !strings.Contains(trimmed, " ") {
		if len(trimmed) <= 3 || isTechnical(trimmed) {
			return QueryTypeKeyword
		}
	}

	// Rule 2: Short query (≤3 words) that looks fully technical → keyword.
	// Skip if query contains stopwords (let Rules 3/4/5 handle those).
	if len(fields) <= 3 {
		technicalCount := 0
		stopwordCount := 0
		for _, f := range fields {
			if isTechnical(f) {
				technicalCount++
			}
			if isStopword(f) {
				stopwordCount++
			}
		}
		// Only classify as keyword if all tokens are technical AND no stopwords.
		if technicalCount >= len(fields) && stopwordCount == 0 {
			return QueryTypeKeyword
		}
	}

	// Rule 3: Very high stopword ratio (>75%) → semantic.
	// Queries like "what is that" (100% stopwords) or "the foo" (50%) are natural language.
	stopwordRatio := stopwordRatio(trimmed, fields)
	if stopwordRatio > 0.75 {
		return QueryTypeSemantic
	}

	// Rule 4: Long natural language phrase (≥7 words) → semantic.
	// Check this before FTS5 operator check so that "and" as a conjunction
	// in natural language queries doesn't incorrectly trigger keyword classification.
	if len(fields) >= 7 {
		return QueryTypeSemantic
	}

	// Rule 5: Check for FTS5 operators → keyword.
	// Only reached for shorter queries where AND/OR/NOT are more likely
	// to be intentional FTS5 operators rather than natural language conjunctions.
	if hasOperators(trimmed) {
		return QueryTypeKeyword
	}

	// Rule 6: Medium-length query (3-6 words) with at least one stopword → hybrid.
	if len(fields) >= 3 && countStopwords(fields) >= 1 {
		return QueryTypeHybrid
	}

	// Rule 7: Short query (2 words) with at least one stopword → hybrid.
	if len(fields) == 2 && countStopwords(fields) >= 1 {
		return QueryTypeHybrid
	}

	// Default fallback: keyword (can still use FTS5).
	return QueryTypeKeyword
}

// hasOperators returns true if the query contains FTS5 boolean operators.
func hasOperators(query string) bool {
	for _, pattern := range operatorPatterns {
		if pattern.MatchString(query) {
			return true
		}
	}
	return false
}

// isTechnical returns true if the token looks like a code/file/technical identifier.
func isTechnical(token string) bool {
	for _, pattern := range technicalPatterns {
		if pattern.MatchString(token) {
			return true
		}
	}
	return false
}

// isStopword returns true if the token is a stopword.
func isStopword(token string) bool {
	lower := strings.ToLower(token)
	for _, sw := range stopwords {
		if lower == sw {
			return true
		}
	}
	return false
}

// stopwordRatio returns the fraction of stopwords in the query.
func stopwordRatio(_ string, fields []string) float64 {
	if len(fields) == 0 {
		return 0
	}
	count := 0
	for _, f := range fields {
		fLower := strings.ToLower(f)
		for _, sw := range stopwords {
			if fLower == sw {
				count++
				break
			}
		}
	}
	return float64(count) / float64(len(fields))
}

// countStopwords returns the number of stopwords in the token list.
func countStopwords(fields []string) int {
	count := 0
	for _, f := range fields {
		fLower := strings.ToLower(f)
		for _, sw := range stopwords {
			if fLower == sw {
				count++
				break
			}
		}
	}
	return count
}

// CountWords returns the number of meaningful words (non-punctuation tokens).
func CountWords(query string) int {
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return 0
	}
	// Filter out pure punctuation tokens.
	count := 0
	for _, f := range fields {
		runes := []rune(f)
		if len(runes) == 0 {
			continue
		}
		// Check if all runes are punctuation or whitespace.
		allPunct := true
		for _, r := range runes {
			if !unicode.IsPunct(r) && !unicode.IsSpace(r) {
				allPunct = false
				break
			}
		}
		if !allPunct {
			count++
		}
	}
	return count
}