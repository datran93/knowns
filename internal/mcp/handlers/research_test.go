package handlers

import (
	"testing"
)

func TestResearchCache(t *testing.T) {
	cache := NewResearchCache()
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	// Test set and get
	cache.set("key1", "value1", "search")
	got, ok := cache.get("key1")
	if !ok || got != "value1" {
		t.Errorf("expected value1, got %q ok=%v", got, ok)
	}

	// Test missing key
	_, ok = cache.get("nonexistent")
	if ok {
		t.Error("expected missing key to return ok=false")
	}

	// Test TTL
	cache.set("key2", "value2", "search")
	got, ok = cache.getWithTTL("key2", 0) // 0 means use default (30 days)
	if !ok || got != "value2" {
		t.Errorf("expected value2, got %q ok=%v", got, ok)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	key := generateCacheKey("search", "test query")
	if key != "search:test query" {
		t.Errorf("expected 'search:test query', got %q", key)
	}
}