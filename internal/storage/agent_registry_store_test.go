package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestAgentRegistryStoreRaceCondition tests TOCTOU in AcquireTaskLock.
// Two agents racing to lock the same task: one should succeed, one should fail.
func TestAgentRegistryStoreRaceCondition(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	// Register two agents.
	agent1 := &AgentSession{ID: "agent-1", PID: 1, Project: root, Name: "test", StartedAt: time.Now()}
	agent2 := &AgentSession{ID: "agent-2", PID: 2, Project: root, Name: "test", StartedAt: time.Now()}
	if err := store.Register(agent1, 5*time.Minute); err != nil {
		t.Fatalf("register agent1: %v", err)
	}
	if err := store.Register(agent2, 5*time.Minute); err != nil {
		t.Fatalf("register agent2: %v", err)
	}

	// Simulate two agents racing to lock the same task concurrently.
	var wg sync.WaitGroup
	results := make([]error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		results[0] = store.AcquireTaskLock("task-X", "agent-1")
	}()
	go func() {
		defer wg.Done()
		results[1] = store.AcquireTaskLock("task-X", "agent-2")
	}()
	wg.Wait()

	// One must succeed, one must fail.
	succeeded := 0
	failed := 0
	for _, r := range results {
		if r == nil {
			succeeded++
		} else {
			failed++
		}
	}

	if succeeded != 1 || failed != 1 {
		t.Errorf("expected 1 success + 1 failure, got %d success + %d failure", succeeded, failed)
	}

	// Verify the lock is held by exactly one agent.
	owner, _ := store.GetTaskLock("task-X")
	if owner == "" {
		t.Error("expected task-X to be locked by one agent after race")
	}
}

// TestAgentRegistryStoreHeartbeat extends TTL.
func TestAgentRegistryStoreHeartbeat(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent := &AgentSession{ID: "agent-hb", PID: 1, Project: root, Name: "test", StartedAt: time.Now()}
	if err := store.Register(agent, 5*time.Minute); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Wait a bit then heartbeat.
	time.Sleep(10 * time.Millisecond)
	if err := store.Heartbeat("agent-hb", 5*time.Minute); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}

	sessions, err := store.ListActive()
	if err != nil {
		t.Fatalf("listActive: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 active session, got %d", len(sessions))
	}
}

// TestAgentRegistryStoreDeleteExpired tests that expired sessions are cleaned up.
func TestAgentRegistryStoreDeleteExpired(t *testing.T) {
	root := t.TempDir()
	store := &AgentRegistryStore{dbPath: filepath.Join(root, "agents.db")}

	// Manually insert an expired session directly in DB to bypass Register's TTL logic.
	if err := store.initDB(); err != nil {
		t.Fatalf("initDB: %v", err)
	}

	db, err := store.db()
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	// Insert an already-expired session.
	expiredTime := time.Now().Add(-1 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO agent_sessions (id, pid, project, name, started_at, heartbeat, expires_at, task_lock)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"agent-expired", 999, root, "expired",
		time.Now(), time.Now(), expiredTime, "")
	if err != nil {
		t.Fatalf("insert expired: %v", err)
	}

	// Insert a valid session.
	validTime := time.Now().Add(5 * time.Minute)
	_, err = db.Exec(`
		INSERT INTO agent_sessions (id, pid, project, name, started_at, heartbeat, expires_at, task_lock)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"agent-valid", 999, root, "valid",
		time.Now(), time.Now(), validTime, "")
	if err != nil {
		t.Fatalf("insert valid: %v", err)
	}

	// DeleteExpired should remove only the expired one.
	if err := store.DeleteExpired(); err != nil {
		t.Fatalf("deleteExpired: %v", err)
	}

	sessions, _ := store.ListActive()
	if len(sessions) != 1 {
		t.Errorf("expected 1 active session after DeleteExpired, got %d", len(sessions))
	}
	if sessions[0].ID != "agent-valid" {
		t.Errorf("expected remaining agent to be agent-valid, got %s", sessions[0].ID)
	}
}

// TestAgentRegistryStoreUnregister removes an agent.
func TestAgentRegistryStoreUnregister(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent := &AgentSession{ID: "agent-remove", PID: 1, Project: root, Name: "test", StartedAt: time.Now()}
	if err := store.Register(agent, 5*time.Minute); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := store.Unregister("agent-remove"); err != nil {
		t.Fatalf("unregister: %v", err)
	}

	sessions, _ := store.ListActive()
	if len(sessions) != 0 {
		t.Errorf("expected 0 active sessions after unregister, got %d", len(sessions))
	}
}

// TestAgentRegistryStoreTaskLockCycle tests full lock acquire/release cycle.
func TestAgentRegistryStoreTaskLockCycle(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent := &AgentSession{ID: "agent-lock", PID: 1, Project: root, Name: "test", StartedAt: time.Now()}
	if err := store.Register(agent, 5*time.Minute); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Acquire lock.
	if err := store.AcquireTaskLock("task-Y", "agent-lock"); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	owner, _ := store.GetTaskLock("task-Y")
	if owner != "agent-lock" {
		t.Errorf("expected task-Y to be locked by agent-lock, got %s", owner)
	}

	// Release lock.
	if err := store.ReleaseTaskLock("task-Y", "agent-lock"); err != nil {
		t.Fatalf("release: %v", err)
	}

	owner, _ = store.GetTaskLock("task-Y")
	if owner != "" {
		t.Errorf("expected task-Y to be unlocked, got %s", owner)
	}
}

// TestAgentRegistryStoreReleaseNonLockedTask does not error when releasing an unlocked task.
func TestAgentRegistryStoreReleaseNonLockedTask(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent := &AgentSession{ID: "agent-x", PID: 1, Project: root, Name: "test", StartedAt: time.Now()}
	if err := store.Register(agent, 5*time.Minute); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Releasing a task that is not locked should not error.
	err := store.ReleaseTaskLock("task-unlocked", "agent-x")
	if err != nil {
		t.Errorf("release of unlocked task should not error, got: %v", err)
	}
}

// TestAgentRegistryStoreReleaseWrongAgent does not error when wrong agent tries to release.
func TestAgentRegistryStoreReleaseWrongAgent(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent1 := &AgentSession{ID: "agent-w1", PID: 1, Project: root, Name: "test1", StartedAt: time.Now()}
	agent2 := &AgentSession{ID: "agent-w2", PID: 2, Project: root, Name: "test2", StartedAt: time.Now()}
	store.Register(agent1, 5*time.Minute)
	store.Register(agent2, 5*time.Minute)

	// Agent1 locks task.
	store.AcquireTaskLock("task-Z", "agent-w1")

	// Agent2 tries to release it — should not error but also not release.
	err := store.ReleaseTaskLock("task-Z", "agent-w2")
	if err != nil {
		t.Errorf("release by wrong agent should not error, got: %v", err)
	}

	// Lock should still be held by agent1.
	owner, _ := store.GetTaskLock("task-Z")
	if owner != "agent-w1" {
		t.Errorf("expected task-Z still locked by agent-w1, got %s", owner)
	}
}

// TestAgentRegistryStoreListActiveEmpty returns nil slice when no agents.
func TestAgentRegistryStoreListActiveEmpty(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	sessions, err := store.ListActive()
	if err != nil {
		t.Fatalf("listActive on empty store: %v", err)
	}
	if sessions == nil {
		// nil slice from empty query result — this is correct SQLite behavior
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

// TestAgentRegistryStoreRegisterCreatesDB creates the DB on first register.
func TestAgentRegistryStoreRegisterCreatesDB(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	// DB does not exist yet.
	if _, err := os.Stat(store.dbPath); err == nil {
		t.Skip("db already exists")
	}

	// Register should create it.
	agent := &AgentSession{ID: "agent-new", PID: 1, Project: root, Name: "new", StartedAt: time.Now()}
	if err := store.Register(agent, 5*time.Minute); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := os.Stat(store.dbPath); err != nil {
		t.Errorf("expected db to be created, stat error: %v", err)
	}
}

// TestAgentRegistryStoreConcurrentHeartbeats is stress test for heartbeat.
func TestAgentRegistryStoreConcurrentHeartbeats(t *testing.T) {
	root := t.TempDir()
	store := NewAgentRegistryStore(root)

	agent := &AgentSession{ID: "agent-stress", PID: 1, Project: root, Name: "stress", StartedAt: time.Now()}
	store.Register(agent, 5*time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Heartbeat("agent-stress", 5*time.Minute)
		}()
	}
	wg.Wait()

	// Should still be active.
	sessions, _ := store.ListActive()
	if len(sessions) != 1 {
		t.Errorf("expected 1 session after concurrent heartbeats, got %d", len(sessions))
	}
}
