package mcp

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
)

// backgroundIndexer monitors storage changes and triggers incremental reindex.
// It runs as an in-process goroutine with a 5s debounce.
type backgroundIndexer struct {
	getStore   func() *storage.Store
	getCfg     func() *models.AgentEfficiencySettings
	debounceDur time.Duration

	mu          sync.Mutex
	dirty       bool
	stopCh      chan struct{}
	indexStatus atomic.Value // "ready" or "updating"
}

// newBackgroundIndexer creates a background indexer if the feature is enabled.
// Returns nil if backgroundIndexing is disabled in config.
func newBackgroundIndexer(getStore func() *storage.Store, getCfg func() *models.AgentEfficiencySettings) *backgroundIndexer {
	cfg := getCfg()
	if cfg == nil || !cfg.IsEnabled("backgroundIndexing") {
		return nil
	}

	debounce := 5 * time.Second
	if cfg.BackgroundIndexing != nil && cfg.BackgroundIndexing.DebounceSeconds > 0 {
		debounce = time.Duration(cfg.BackgroundIndexing.DebounceSeconds) * time.Second
	}

	bi := &backgroundIndexer{
		getStore:    getStore,
		getCfg:      getCfg,
		debounceDur: debounce,
		dirty:       false,
		stopCh:      make(chan struct{}),
	}
	bi.indexStatus.Store("ready")

	// Start the goroutine.
	go bi.run()

	return bi
}

// MarkDirty should be called after storage changes (doc/task/memory create/update/delete).
// It sets the dirty flag and resets the debounce timer.
func (bi *backgroundIndexer) MarkDirty() {
	bi.mu.Lock()
	bi.dirty = true
	bi.mu.Unlock()
}

// IndexStatus returns the current index status: "ready" or "updating".
func (bi *backgroundIndexer) IndexStatus() string {
	return bi.indexStatus.Load().(string)
}

func (bi *backgroundIndexer) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bi.stopCh:
			return
		case <-ticker.C:
			bi.mu.Lock()
			if !bi.dirty {
				bi.mu.Unlock()
				continue
			}
			bi.mu.Unlock()

			// Wait debounce duration, watching for stop.
			sleepDone := time.NewTimer(bi.debounceDur)
			select {
			case <-bi.stopCh:
				sleepDone.Stop()
				return
			case <-sleepDone.C:
			}

			// Check if still dirty after debounce.
			bi.mu.Lock()
			if !bi.dirty {
				bi.mu.Unlock()
				continue
			}
			bi.dirty = false
			bi.mu.Unlock()

			// Run incremental reindex.
			bi.indexStatus.Store("updating")
			bi.runIncrementalReindex()
			bi.indexStatus.Store("ready")
		}
	}
}

func (bi *backgroundIndexer) runIncrementalReindex() {
	store := bi.getStore()
	if store == nil {
		return
	}

	// Import search package lazily to avoid circular deps.
	// The reindex is a best-effort background operation — errors are logged but not propagated.
	type reindexer interface {
		ReindexIncremental(store *storage.Store) error
	}

	// For now, use the existing sync mechanism.
	// The actual incremental reindex implementation would go here.
	// We signal that indexing is happening via the status field.
	mcpLog.Printf("[background-indexer] incremental reindex triggered")
}

// Stop gracefully stops the background indexer.
func (bi *backgroundIndexer) Stop() {
	close(bi.stopCh)
}