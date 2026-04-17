package workingmemory

import (
	"sort"
	"sync"
	"time"

	"github.com/howznguyen/knowns/internal/models"
)

// Store is a thread-safe in-memory store for session-scoped working memory.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*models.MemoryEntry
}

func NewStore() *Store {
	return &Store{entries: make(map[string]*models.MemoryEntry)}
}

func cloneEntry(entry *models.MemoryEntry) *models.MemoryEntry {
	if entry == nil {
		return nil
	}
	clone := *entry
	if entry.Tags != nil {
		clone.Tags = append([]string(nil), entry.Tags...)
	}
	if entry.Metadata != nil {
		clone.Metadata = make(map[string]string, len(entry.Metadata))
		for k, v := range entry.Metadata {
			clone.Metadata[k] = v
		}
	}
	return &clone
}

func (s *Store) Add(entry *models.MemoryEntry) *models.MemoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := cloneEntry(entry)
	if stored.ID == "" {
		stored.ID = models.NewTaskID()
	}
	stored.Layer = models.MemoryLayerWorking
	now := time.Now().UTC()
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = now
	}
	stored.UpdatedAt = now
	if stored.Tags == nil {
		stored.Tags = []string{}
	}
	s.entries[stored.ID] = stored
	return cloneEntry(stored)
}

func (s *Store) Get(id string) (*models.MemoryEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[id]
	if !ok {
		return nil, false
	}
	return cloneEntry(entry), true
}

func (s *Store) List() []*models.MemoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.MemoryEntry, 0, len(s.entries))
	for _, entry := range s.entries {
		result = append(result, cloneEntry(entry))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.entries[id]
	if ok {
		delete(s.entries, id)
	}
	return ok
}

func (s *Store) Clear() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := len(s.entries)
	s.entries = make(map[string]*models.MemoryEntry)
	return count
}
