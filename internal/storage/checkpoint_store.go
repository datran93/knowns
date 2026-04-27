package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionCheckpoint stores agent session state for resume.
type SessionCheckpoint struct {
	ID              string    `json:"id"`
	LastTaskID      string    `json:"lastTaskId"`
	RecentDocs      []string  `json:"recentDocs"`       // last 10 doc paths
	RecentTasks     []string  `json:"recentTasks"`     // last 10 task IDs
	PendingDecisions []string `json:"pendingDecisions"`
	CodeGraphSummary string   `json:"codeGraphSummary"` // summary of indexed symbols
	CreatedAt       time.Time `json:"createdAt"`
	ExpiresAt       time.Time `json:"expiresAt"` // TTL based on config
}

// SessionCheckpointStore manages session checkpoints in .knowns/sessions/
type SessionCheckpointStore struct {
	root     string
	ttlHours int
}

// NewSessionCheckpointStore creates a checkpoint store rooted at root/sessions/.
// Default TTL is ttlHours (used when TTL hasn't been set via UpdateTTL).
func NewSessionCheckpointStore(root string, ttlHours int) *SessionCheckpointStore {
	return &SessionCheckpointStore{
		root:     filepath.Join(root, "sessions"),
		ttlHours: ttlHours,
	}
}

// UpdateTTL updates the TTL used for new checkpoints.
func (cs *SessionCheckpointStore) UpdateTTL(ttlHours int) {
	cs.ttlHours = ttlHours
}

// sessionsDir returns the sessions directory path.
func (cs *SessionCheckpointStore) sessionsDir() string {
	return cs.root
}

// Save writes a checkpoint to sessions/<id>.json atomically.
func (cs *SessionCheckpointStore) Save(cp *SessionCheckpoint) error {
	if err := os.MkdirAll(cs.sessionsDir(), 0755); err != nil {
		return fmt.Errorf("checkpoint save: mkdir: %w", err)
	}

	// Set expiration based on current TTL.
	cp.ExpiresAt = time.Now().Add(time.Duration(cs.ttlHours) * time.Hour)

	path := filepath.Join(cs.sessionsDir(), cp.ID+".json")
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint save: marshal: %w", err)
	}
	data = append(data, '\n')
	return atomicWrite(path, data)
}

// LoadLatest returns the most recent non-expired checkpoint.
// Returns nil, nil if no valid checkpoint exists.
func (cs *SessionCheckpointStore) LoadLatest() (*SessionCheckpoint, error) {
	entries, err := os.ReadDir(cs.sessionsDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checkpoint loadlatest: readdir: %w", err)
	}

	var checkpoints []*SessionCheckpoint
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		// ID is filename without .json
		id := strings.TrimSuffix(e.Name(), ".json")
		cp, err := cs.Load(id)
		if err != nil {
			continue
		}
		if time.Now().Before(cp.ExpiresAt) {
			checkpoints = append(checkpoints, cp)
		}
	}

	if len(checkpoints) == 0 {
		return nil, nil
	}

	// Sort by CreatedAt descending.
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].CreatedAt.After(checkpoints[j].CreatedAt)
	})
	return checkpoints[0], nil
}

// Load returns a specific checkpoint by ID.
func (cs *SessionCheckpointStore) Load(id string) (*SessionCheckpoint, error) {
	path := filepath.Join(cs.sessionsDir(), id+".json")
	var cp SessionCheckpoint
	if err := readJSON(path, &cp); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint %q not found", id)
		}
		return nil, fmt.Errorf("checkpoint load: %w", err)
	}
	return &cp, nil
}

// Delete removes a checkpoint by ID.
func (cs *SessionCheckpointStore) Delete(id string) error {
	path := filepath.Join(cs.sessionsDir(), id+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("checkpoint %q not found", id)
		}
		return fmt.Errorf("checkpoint delete: %w", err)
	}
	return nil
}

// List returns all checkpoints (including expired ones).
func (cs *SessionCheckpointStore) List() ([]*SessionCheckpoint, error) {
	entries, err := os.ReadDir(cs.sessionsDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checkpoint list: readdir: %w", err)
	}

	var checkpoints []*SessionCheckpoint
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		cp, err := cs.Load(id)
		if err != nil {
			continue
		}
		checkpoints = append(checkpoints, cp)
	}
	return checkpoints, nil
}

// DeleteExpired removes all expired checkpoints.
func (cs *SessionCheckpointStore) DeleteExpired() error {
	checkpoints, err := cs.List()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, cp := range checkpoints {
		if now.After(cp.ExpiresAt) {
			if err := cs.Delete(cp.ID); err != nil {
				// Log but continue.
				continue
			}
		}
	}
	return nil
}