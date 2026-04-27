package mcp

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
)

var agentIDCounter int64

// agentRegistry handles multi-agent awareness and task locking.
// Note: Start() and Stop() must be called to activate the heartbeat goroutine.
type agentRegistry struct {
	getStore func() *storage.Store
	getCfg   func() *models.AgentEfficiencySettings

	mu      sync.Mutex
	stopCh  chan struct{}
	stopped bool
	agentID string
}

// newAgentRegistry creates an agent registry if multiAgent is enabled.
// Returns nil if the feature is disabled.
func newAgentRegistry(getStore func() *storage.Store, getCfg func() *models.AgentEfficiencySettings) *agentRegistry {
	cfg := getCfg()
	if cfg == nil || !cfg.IsEnabled("multiAgent") {
		return nil
	}

	ar := &agentRegistry{
		getStore: getStore,
		getCfg:   getCfg,
		stopCh:   make(chan struct{}),
	}

	ar.agentID = fmt.Sprintf("agent-%d-%d", os.Getpid(), atomic.AddInt64(&agentIDCounter, 1))
	return ar
}

// Start registers the agent and starts the heartbeat goroutine.
func (ar *agentRegistry) Start() {
	ar.registerAgent()
	go ar.heartbeatLoop()
}

// Stop unregisters the agent gracefully.
func (ar *agentRegistry) Stop() {
	ar.mu.Lock()
	if ar.stopped {
		ar.mu.Unlock()
		return
	}
	ar.stopped = true
	ar.mu.Unlock()
	close(ar.stopCh)

	store := ar.getStore()
	if store != nil && store.Agents != nil {
		store.Agents.Unregister(ar.agentID)
	}
}

// AgentID returns the unique agent ID for this registry instance.
func (ar *agentRegistry) AgentID() string {
	return ar.agentID
}

// registerAgent registers the agent session with the registry.
func (ar *agentRegistry) registerAgent() {
	store := ar.getStore()
	if store == nil || store.Agents == nil {
		return
	}

	cfg := ar.getCfg()
	ttl := 5 * time.Minute
	if cfg != nil && cfg.MultiAgent != nil && cfg.MultiAgent.LockTTLSeconds > 0 {
		ttl = time.Duration(cfg.MultiAgent.LockTTLSeconds) * time.Second
	}

	session := &storage.AgentSession{
		ID:        ar.agentID,
		PID:       os.Getpid(),
		Project:   store.Root,
		Name:      "knowns-mcp",
		StartedAt: time.Now(),
	}
	if err := store.Agents.Register(session, ttl); err != nil {
		mcpLog.Printf("[agent-registry] register failed: %v", err)
	}
}

// heartbeatLoop sends periodic heartbeats until stopped.
func (ar *agentRegistry) heartbeatLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ar.stopCh:
			return
		case <-ticker.C:
			store := ar.getStore()
			if store != nil && store.Agents != nil {
				cfg := ar.getCfg()
				ttl := 5 * time.Minute
				if cfg != nil && cfg.MultiAgent != nil && cfg.MultiAgent.LockTTLSeconds > 0 {
					ttl = time.Duration(cfg.MultiAgent.LockTTLSeconds) * time.Second
				}
				store.Agents.Heartbeat(ar.agentID, ttl)
			}
		}
	}
}
