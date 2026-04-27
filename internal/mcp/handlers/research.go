package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ResearchCache implements a simple in-memory TTL cache for research results.
type ResearchCache struct {
	mu      sync.RWMutex
	data    map[string]cacheEntry
	maxAge  time.Duration
}

type cacheEntry struct {
	content     string
	fetchedAt   time.Time
	contentType string // "search" or "url_full"
}

func NewResearchCache() *ResearchCache {
	return &ResearchCache{
		data:   make(map[string]cacheEntry),
		maxAge: 30 * 24 * time.Hour, // 30 days default
	}
}

func (c *ResearchCache) get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.data[key]
	if !ok {
		return "", false
	}
	if time.Since(entry.fetchedAt) > c.maxAge {
		return "", false
	}
	return entry.content, true
}

func (c *ResearchCache) set(key, content, contentType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = cacheEntry{
		content:     content,
		fetchedAt:   time.Now(),
		contentType: contentType,
	}
}

func (c *ResearchCache) getWithTTL(key string, ttlHours int) (string, bool) {
	if ttlHours <= 0 {
		return c.get(key)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.data[key]
	if !ok {
		return "", false
	}
	maxAge := time.Duration(ttlHours) * time.Hour
	if time.Since(entry.fetchedAt) > maxAge {
		return "", false
	}
	return entry.content, true
}

// SearchResult holds a single search result.
type SearchResult struct {
	Title string `json:"title"`
	Href  string `json:"href"`
	Body  string `json:"body"`
}

// SearchWithFallback attempts to fetch search results, trying multiple backends.
func SearchWithFallback(query string, maxResults int) ([]SearchResult, error) {
	// Try DuckDuckGo HTML scraper first
	results, err := searchDuckDuckGo(query, maxResults)
	if err == nil && len(results) > 0 {
		return results, nil
	}
	// Fallback to a simple Google-style search (using serpapi or direct)
	return results, fmt.Errorf("search failed: no results")
}

// searchDuckDuckGo performs a DuckDuckGo search.
func searchDuckDuckGo(query string, maxResults int) ([]SearchResult, error) {
	url := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", strings.ReplaceAll(query, " ", "+"))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseDuckDuckGoResults(string(body), maxResults)
}

// parseDuckDuckGoResults extracts search results from DuckDuckGo HTML.
func parseDuckDuckGoResults(html string, maxResults int) ([]SearchResult, error) {
	var results []SearchResult
	// Simple HTML parsing - in production use a proper HTML parser
	lines := strings.Split(html, "\n")
	for _, line := range lines {
		if strings.Contains(line, "result__a") || strings.Contains(line, "class=\"a\"") {
			// Extract URL
			idx := strings.Index(line, "href=\"")
			if idx == -1 {
				continue
			}
			idx += 6
			endIdx := strings.Index(line[idx:], "\"")
			if endIdx == -1 {
				continue
			}
			url := line[idx : idx+endIdx]
			if !strings.HasPrefix(url, "http") {
				continue
			}
			// Try to extract title from the same or nearby lines
			title := extractTitleFromLine(line)
			results = append(results, SearchResult{
				Title: title,
				Href:  url,
				Body:  "Click to view",
			})
			if len(results) >= maxResults {
				break
			}
		}
	}
	return results, nil
}

func extractTitleFromLine(line string) string {
	// Try to find the text between > and </a>
	parts := strings.Split(line, ">")
	for i := 1; i < len(parts); i++ {
		if strings.Contains(parts[i], "</a>") {
			title := strings.Split(parts[i], "</a>")[0]
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}
	return "Untitled"
}

// FetchJinaMarkdown fetches a URL and converts it to markdown using Jina AI reader.
func FetchJinaMarkdown(url string, hint string) string {
	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", url)
	req, _ := http.NewRequest("GET", jinaURL, nil)
	req.Header.Set("Accept", "text/markdown")
	if hint != "" {
		req.Header.Set("X-Return-Format", "markdown")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error fetching markdown: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}
	return string(body)
}

// generateCacheKey creates a cache key for search queries.
func generateCacheKey(prefix, query string) string {
	return prefix + ":" + query
}

// researchCache is a shared cache instance.
var researchCache = NewResearchCache()

// RegisterResearchTool registers the doc research MCP tool.
func RegisterResearchTool(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("research",
			mcp.WithDescription("Research operations for real-time web search and documentation. Use 'action' to specify: search_latest_syntax."),
			mcp.WithString("action",
				mcp.Required(),
				mcp.Description("Action to perform"),
				mcp.Enum("search_latest_syntax"),
			),
			mcp.WithString("topic",
				mcp.Description("The specific concept to research (e.g., 'React server components', 'Next.js 14 App Router')"),
			),
			mcp.WithArray("libraries",
				mcp.Description("Optional list of specific libraries to narrow context"),
				mcp.WithStringItems(),
			),
			mcp.WithNumber("max_age_hours",
				mcp.Description("Cache TTL override in hours (default: 720 = 30 days)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return errResult("action is required")
			}
			switch action {
			case "search_latest_syntax":
				return handleSearchLatestSyntax(req)
			default:
				return errResultf("unknown research action: %s", action)
			}
		},
	)
}

func handleSearchLatestSyntax(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	topic, err := req.RequireString("topic")
	if err != nil {
		return errResult("topic is required")
	}

	searchQuery := topic
	if libsRaw, ok := args["libraries"].([]interface{}); ok && len(libsRaw) > 0 {
		var libs []string
		for _, l := range libsRaw {
			if s, ok := l.(string); ok {
				libs = append(libs, s)
			}
		}
		if len(libs) > 0 {
			searchQuery += " " + strings.Join(libs, " ")
		}
	}

	maxAgeHours := 0
	if v, ok := args["max_age_hours"].(float64); ok && v > 0 {
		maxAgeHours = int(v)
	}

	cacheKey := generateCacheKey("search", searchQuery)
	if cachedResult, ok := researchCache.getWithTTL(cacheKey, maxAgeHours); ok {
		return mcp.NewToolResultText(fmt.Sprintf("⚡ (Cached - Loaded instantly from memory)\n%s", cachedResult)), nil
	}

	results, err := SearchWithFallback(searchQuery+" tutorial OR documentation", 3)
	if err != nil || len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("❌ No recent results found for: %s", topic)), nil
	}

	var finalReport strings.Builder
	finalReport.WriteString(fmt.Sprintf("🔍 REAL-TIME RESEARCH RESULTS FOR: '%s'\n\n", topic))
	finalReport.WriteString("### 1. QUICK SNIPPETS (SEARCH ENGINE RESULTS)\n")

	for i, res := range results {
		finalReport.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, res.Title, res.Href))
		finalReport.WriteString(fmt.Sprintf("   Snippet: %s\n\n", res.Body))
	}

	topURL := results[0].Href
	if topURL != "" {
		finalReport.WriteString("### 2. DEEP DIVE (PARTIAL EXTRACTION OF TOP RESULT)\n")
		finalReport.WriteString(fmt.Sprintf("Source: %s\n", topURL))

		cacheKeyFull := generateCacheKey("url_full", topURL)
		var fullContent string
		if cachedFull, ok := researchCache.get(cacheKeyFull); ok {
			fullContent = cachedFull
		} else {
			fullContent = FetchJinaMarkdown(topURL, "")
			if !strings.HasPrefix(fullContent, "Error fetching markdown") {
				researchCache.set(cacheKeyFull, fullContent, "url_full")
			}
		}

		truncatedLen := 4000
		if len(fullContent) < 4000 {
			truncatedLen = len(fullContent)
		}
		truncated := fullContent[:truncatedLen]

		if len(fullContent) > 4000 {
			finalReport.WriteString(fmt.Sprintf("Reading content... (Previewing first %d/%d characters)\n", truncatedLen, len(fullContent)))
		}
		finalReport.WriteString(truncated)
		if len(fullContent) > 4000 {
			finalReport.WriteString("\n\n... (content truncated, use direct URL for full)")
		}
	}

	result := finalReport.String()
	researchCache.set(cacheKey, result, "search")
	return mcp.NewToolResultText(result), nil
}