package main

import (
	"fmt"
	"log"
	"sync"
)

// SessionManager manages multiple bot sessions
type SessionManager struct {
	sessions      map[string]*Session
	mutex         sync.RWMutex
	nextPort      int
	nextSessionID int
}

// NewSessionManager creates a new session manager
func NewSessionManager(startPort int) *SessionManager {
	return &SessionManager{
		sessions:      make(map[string]*Session),
		nextPort:      startPort,
		nextSessionID: 1,
	}
}

// AddSession creates and starts a new session with auto-generated ID
// This method blocks until the user completes the Steam login
func (sm *SessionManager) AddSession(userID string) (*Session, error) {
	sm.mutex.Lock()

	// Auto-generate session ID
	id := fmt.Sprintf("session-%d", sm.nextSessionID)
	sm.nextSessionID++

	port := sm.nextPort
	sm.nextPort++

	session := NewSession(id, userID, port)

	sm.mutex.Unlock()

	log.Printf("Creating session %s for user %s on port %d", id, userID, port)
	log.Printf("Browser will open - please complete Steam login...")

	// This blocks until login is complete
	if err := session.Start(); err != nil {
		return nil, err
	}

	sm.mutex.Lock()
	sm.sessions[id] = session
	sm.mutex.Unlock()

	log.Printf("✓ Session %s is now fully logged in and ready!", id)

	return session, nil
}

// RemoveSession stops and removes a session
func (sm *SessionManager) RemoveSession(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	if err := session.Stop(); err != nil {
		return err
	}

	delete(sm.sessions, id)
	log.Printf("Session %s removed", id)

	return nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// StopAll stops all sessions
func (sm *SessionManager) StopAll() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for id, session := range sm.sessions {
		if err := session.Stop(); err != nil {
			log.Printf("Error stopping session %s: %v", id, err)
		}
	}

	sm.sessions = make(map[string]*Session)
	log.Println("All sessions stopped")
}
