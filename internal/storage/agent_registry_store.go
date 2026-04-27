package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AgentSession represents an active agent session.
type AgentSession struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Project   string    `json:"project"` // project root
	Name      string    `json:"name"`    // agent name (e.g., "claude", "cursor")
	StartedAt time.Time `json:"startedAt"`
	Heartbeat time.Time `json:"heartbeat"`
	ExpiresAt time.Time `json:"expiresAt"` // TTL based on lock TTL config
	TaskLock  string    `json:"taskLock,omitempty"` // locked task ID (if any)
}

// AgentRegistryStore manages active agent sessions in SQLite.
type AgentRegistryStore struct {
	dbPath string
}

// NewAgentRegistryStore creates a store at root/agents.db.
func NewAgentRegistryStore(root string) *AgentRegistryStore {
	return &AgentRegistryStore{
		dbPath: filepath.Join(root, "agents.db"),
	}
}

func (s *AgentRegistryStore) db() (*sql.DB, error) {
	return sql.Open("sqlite", s.dbPath)
}

func (s *AgentRegistryStore) initDB() error {
	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_sessions (
			id TEXT PRIMARY KEY,
			pid INTEGER NOT NULL,
			project TEXT NOT NULL,
			name TEXT NOT NULL,
			started_at TEXT NOT NULL,
			heartbeat TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			task_lock TEXT DEFAULT ''
		)
	`)
	return err
}

func (s *AgentRegistryStore) ensureDB() error {
	if _, err := os.Stat(s.dbPath); os.IsNotExist(err) {
		return s.initDB()
	}
	return nil
}

// Register adds or updates an agent session with the given TTL.
func (s *AgentRegistryStore) Register(session *AgentSession, ttl time.Duration) error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	session.Heartbeat = time.Now()
	session.ExpiresAt = time.Now().Add(ttl)

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO agent_sessions (id, pid, project, name, started_at, heartbeat, expires_at, task_lock)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			pid = excluded.pid,
			project = excluded.project,
			name = excluded.name,
			heartbeat = excluded.heartbeat,
			expires_at = excluded.expires_at,
			task_lock = excluded.task_lock
	`,
		session.ID, session.PID, session.Project, session.Name,
		session.StartedAt, session.Heartbeat, session.ExpiresAt, session.TaskLock)

	return err
}

// Heartbeat updates the heartbeat timestamp and extends TTL.
func (s *AgentRegistryStore) Heartbeat(agentID string, ttl time.Duration) error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	expiresAt := time.Now().Add(ttl)
	_, err = db.Exec(`
		UPDATE agent_sessions SET heartbeat = ?, expires_at = ? WHERE id = ?`,
		time.Now(), expiresAt, agentID)
	return err
}

// Unregister removes an agent session.
func (s *AgentRegistryStore) Unregister(agentID string) error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM agent_sessions WHERE id = ?`, agentID)
	return err
}

// ListActive returns all non-expired active sessions.
func (s *AgentRegistryStore) ListActive() ([]*AgentSession, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}

	db, err := s.db()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	now := time.Now()
	rows, err := db.Query(`
		SELECT id, pid, project, name, started_at, heartbeat, expires_at, task_lock
		FROM agent_sessions
		WHERE expires_at > ?
		ORDER BY started_at DESC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*AgentSession
	for rows.Next() {
		var ses AgentSession
		var startedAt, heartbeat, expiresAt string
		if err := rows.Scan(&ses.ID, &ses.PID, &ses.Project, &ses.Name, &startedAt, &heartbeat, &expiresAt, &ses.TaskLock); err != nil {
			continue
		}
		ses.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		ses.Heartbeat, _ = time.Parse(time.RFC3339, heartbeat)
		ses.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
		sessions = append(sessions, &ses)
	}
	return sessions, nil
}

// GetTaskLock returns the agent ID holding a lock on the given task.
func (s *AgentRegistryStore) GetTaskLock(taskID string) (string, error) {
	if err := s.ensureDB(); err != nil {
		return "", err
	}

	db, err := s.db()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var agentID string
	var expiresAt string
	err = db.QueryRow(`
		SELECT id, expires_at FROM agent_sessions WHERE task_lock = ? AND expires_at > ?`,
		taskID, time.Now()).Scan(&agentID, &expiresAt)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return agentID, err
}

// AcquireTaskLock attempts to lock a task for an agent. Returns error if locked by another.
func (s *AgentRegistryStore) AcquireTaskLock(taskID, agentID string) error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	// First check if any other agent holds the lock.
	owner, err := s.GetTaskLock(taskID)
	if err != nil {
		return err
	}
	if owner != "" && owner != agentID {
		return fmt.Errorf("task %q is locked by agent %q", taskID, owner)
	}

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE agent_sessions SET task_lock = ? WHERE id = ?`, taskID, agentID)
	return err
}

// ReleaseTaskLock releases a task lock held by an agent.
func (s *AgentRegistryStore) ReleaseTaskLock(taskID, agentID string) error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE agent_sessions SET task_lock = '' WHERE id = ? AND task_lock = ?`, agentID, taskID)
	return err
}

// DeleteExpired removes expired sessions and releases their task locks.
func (s *AgentRegistryStore) DeleteExpired() error {
	if err := s.ensureDB(); err != nil {
		return err
	}

	db, err := s.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM agent_sessions WHERE expires_at <= ?`, time.Now())
	return err
}
