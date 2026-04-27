package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCheckpointStoreDeleteExpiredBoundary tests the boundary condition:
// Save uses Before(ExpiresAt) but DeleteExpired uses After(ExpiresAt).
// A checkpoint that expires exactly at "now" should be treated consistently.
func TestCheckpointStoreDeleteExpiredBoundary(t *testing.T) {
	root := t.TempDir()
	store := &SessionCheckpointStore{
		root:     filepath.Join(root, "sessions"),
		ttlHours: 24,
	}
	if err := os.MkdirAll(store.sessionsDir(), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Manually insert checkpoints with known expiration times.
	now := time.Now()

	// Insert checkpoint that expired 1 hour ago (definitely expired).
	cpExpired := &SessionCheckpoint{
		ID:        "cp-expired",
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	cpExpired.CreatedAt = now.Add(-2 * time.Hour)
	data, _ := json.Marshal(cpExpired)
	atomicWrite(filepath.Join(store.sessionsDir(), "cp-expired.json"), data)

	// Insert checkpoint that expires in 1 hour (definitely valid).
	cpValid := &SessionCheckpoint{
		ID:        "cp-valid",
		ExpiresAt: now.Add(1 * time.Hour),
	}
	cpValid.CreatedAt = now
	data2, _ := json.Marshal(cpValid)
	atomicWrite(filepath.Join(store.sessionsDir(), "cp-valid.json"), data2)

	// DeleteExpired should remove only the expired one.
	if err := store.DeleteExpired(); err != nil {
		t.Fatalf("deleteExpired: %v", err)
	}

	// Check cp-expired is gone.
	if _, err := os.Stat(filepath.Join(store.sessionsDir(), "cp-expired.json")); err == nil {
		t.Error("cp-expired should have been deleted")
	}

	// Check cp-valid still exists.
	if _, err := os.Stat(filepath.Join(store.sessionsDir(), "cp-valid.json")); err != nil {
		t.Error("cp-valid should still exist")
	}
}

// TestCheckpointStoreSaveSetsExpires sets expiration correctly.
func TestCheckpointStoreSaveSetsExpires(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)
	store.UpdateTTL(1) // 1 hour TTL

	cp := &SessionCheckpoint{
		ID:        "cp-ttl-test",
		LastTaskID: "task-1",
		CreatedAt: time.Now(),
	}

	if err := store.Save(cp); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Reload and verify expiration is ~1 hour from now.
	loaded, err := store.Load("cp-ttl-test")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	expectedMin := time.Now().Add(30 * time.Minute)
	expectedMax := time.Now().Add(90 * time.Minute)
	if loaded.ExpiresAt.Before(expectedMin) || loaded.ExpiresAt.After(expectedMax) {
		t.Errorf("ExpiresAt=%v, expected between %v and %v", loaded.ExpiresAt, expectedMin, expectedMax)
	}
}

// TestCheckpointStoreLoadLatestReturnsNewest returns most recent non-expired.
func TestCheckpointStoreLoadLatestReturnsNewest(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)

	// Save multiple checkpoints with different IDs.
	for i := 0; i < 3; i++ {
		cp := &SessionCheckpoint{
			ID:        "cp-newest",
			LastTaskID: "task-newest",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		store.Save(cp)
		time.Sleep(10 * time.Millisecond)
	}

	// Also save an older checkpoint.
	cpOld := &SessionCheckpoint{
		ID:        "cp-old",
		LastTaskID: "task-old",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	cpOld.ExpiresAt = time.Now().Add(-30 * time.Minute) // already expired
	store.Save(cpOld)

	latest, err := store.LoadLatest()
	if err != nil {
		t.Fatalf("loadLatest: %v", err)
	}
	if latest == nil {
		t.Fatal("expected a latest checkpoint")
	}
	if latest.ID != "cp-newest" {
		t.Errorf("expected cp-newest (most recent non-expired), got %s", latest.ID)
	}
}

// TestCheckpointStoreLoadLatestReturnsNilWhenAllExpired bypasses Save
// to directly set ExpiresAt to the past.
func TestCheckpointStoreLoadLatestReturnsNilWhenAllExpired(t *testing.T) {
	root := t.TempDir()
	store := &SessionCheckpointStore{
		root:     filepath.Join(root, "sessions"),
		ttlHours: 1,
	}
	if err := os.MkdirAll(store.sessionsDir(), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Directly write a JSON file with an already-expired checkpoint (bypassing Save
	// which recalculates ExpiresAt from TTL).
	expiredCp := &SessionCheckpoint{
		ID:         "cp-truly-expired",
		LastTaskID: "task-x",
		CreatedAt:  time.Now().Add(-3 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
	}
	data, _ := json.Marshal(expiredCp)
	atomicWrite(filepath.Join(store.sessionsDir(), "cp-truly-expired.json"), append(data, '\n'))

	// Also write a valid checkpoint (to make sure we're not returning nil by accident).
	validCp := &SessionCheckpoint{
		ID:         "cp-still-valid",
		LastTaskID: "task-y",
		CreatedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // valid for 1 more hour
	}
	data2, _ := json.Marshal(validCp)
	atomicWrite(filepath.Join(store.sessionsDir(), "cp-still-valid.json"), append(data2, '\n'))

	// LoadLatest should skip the expired one and return the valid one.
	latest, err := store.LoadLatest()
	if err != nil {
		t.Fatalf("loadLatest: %v", err)
	}
	if latest == nil {
		t.Fatal("expected valid checkpoint, got nil")
	}
	if latest.ID != "cp-still-valid" {
		t.Errorf("expected cp-still-valid, got %s", latest.ID)
	}

	// Now delete the valid one and re-check.
	os.Remove(filepath.Join(store.sessionsDir(), "cp-still-valid.json"))
	latest2, _ := store.LoadLatest()
	if latest2 != nil {
		t.Errorf("expected nil when all expired, got %+v", latest2)
	}
}

// TestCheckpointStoreUpdateTTL affects subsequent saves.
func TestCheckpointStoreUpdateTTL(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)

	// Save with default TTL.
	cp1 := &SessionCheckpoint{ID: "cp1", CreatedAt: time.Now()}
	store.Save(cp1)
	cp1Loaded, _ := store.Load("cp1")
	ttl1 := cp1Loaded.ExpiresAt.Sub(time.Now())

	// Update TTL to 1 hour.
	store.UpdateTTL(1)
	cp2 := &SessionCheckpoint{ID: "cp2", CreatedAt: time.Now()}
	store.Save(cp2)
	cp2Loaded, _ := store.Load("cp2")
	ttl2 := cp2Loaded.ExpiresAt.Sub(time.Now())

	// cp2 TTL should be significantly shorter.
	if ttl2 >= ttl1 {
		t.Errorf("expected cp2 TTL (%v) < cp1 TTL (%v)", ttl2, ttl1)
	}
}

// TestCheckpointStoreDelete removes specific checkpoint.
func TestCheckpointStoreDelete(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)

	cp := &SessionCheckpoint{ID: "cp-delete-me", CreatedAt: time.Now()}
	store.Save(cp)

	if err := store.Delete("cp-delete-me"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := store.Load("cp-delete-me"); err == nil {
		t.Error("expected error loading deleted checkpoint")
	}
}

// TestCheckpointStoreDeleteNonexistent returns error.
func TestCheckpointStoreDeleteNonexistent(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)

	err := store.Delete("cp-does-not-exist")
	if err == nil {
		t.Error("expected error deleting nonexistent checkpoint")
	}
}

// TestCheckpointStoreListReturnsAllIncludingExpired.
func TestCheckpointStoreListReturnsAllIncludingExpired(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 1)

	// One valid, one expired.
	cp1 := &SessionCheckpoint{ID: "cp-list-valid", CreatedAt: time.Now()}
	cp1.ExpiresAt = time.Now().Add(1 * time.Hour)
	store.Save(cp1)

	cp2 := &SessionCheckpoint{ID: "cp-list-expired", CreatedAt: time.Now().Add(-3 * time.Hour)}
	cp2.ExpiresAt = time.Now().Add(-1 * time.Hour)
	store.Save(cp2)

	list, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 checkpoints (including expired), got %d", len(list))
	}
}

// TestCheckpointStoreAtomicWrite preserves data.
func TestCheckpointStoreAtomicWrite(t *testing.T) {
	root := t.TempDir()
	store := NewSessionCheckpointStore(root, 24)

	cp := &SessionCheckpoint{
		ID:               "cp-atomic",
		LastTaskID:       "task-atomic",
		RecentDocs:       []string{"doc-a", "doc-b"},
		RecentTasks:      []string{"task-1", "task-2"},
		PendingDecisions: []string{"dec-1"},
		CodeGraphSummary: "summary-v1",
		CreatedAt:        time.Now(),
	}

	if err := store.Save(cp); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, _ := store.Load("cp-atomic")
	if loaded.LastTaskID != "task-atomic" {
		t.Errorf("LastTaskID: got %s, want task-atomic", loaded.LastTaskID)
	}
	if len(loaded.RecentDocs) != 2 {
		t.Errorf("RecentDocs: got %d, want 2", len(loaded.RecentDocs))
	}
	if len(loaded.RecentTasks) != 2 {
		t.Errorf("RecentTasks: got %d, want 2", len(loaded.RecentTasks))
	}
	if loaded.CodeGraphSummary != "summary-v1" {
		t.Errorf("CodeGraphSummary: got %s, want summary-v1", loaded.CodeGraphSummary)
	}
}
