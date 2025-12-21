package bot

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// SessionManager manages multiple bot sessions
type SessionManager struct {
	sessions      map[string]*SessionChrome
	mutex         sync.RWMutex
	nextPort      int
	nextSessionID int
	nextDebugPort int
	timeouts      map[string]*time.Timer
	credStore     *CredentialStore
}

// NewSessionManager creates a new session manager
func NewSessionManager(startPort int) *SessionManager {
	// Initialize credential store
	credStore, err := NewCredentialStore("steam_credentials.json")
	if err != nil {
		log.Printf("Warning: Failed to load credentials: %v", err)
	}

	return &SessionManager{
		sessions:      make(map[string]*SessionChrome),
		nextPort:      startPort,
		nextSessionID: 1,
		nextDebugPort: 9222,
		timeouts:      make(map[string]*time.Timer),
		credStore:     credStore,
	}
}

// AddSession creates and starts a new session with auto-generated ID
// This method blocks until the user completes the Steam login
func (sm *SessionManager) AddSession(userID string) (*SessionChrome, error) {
	sm.mutex.Lock()

	// Auto-generate session ID
	id := fmt.Sprintf("session-%d", sm.nextSessionID)
	sm.nextSessionID++

	port := sm.nextPort
	sm.nextPort++

	debugPort := sm.nextDebugPort
	sm.nextDebugPort++

	// Check if we have credentials for this account
	var session *SessionChrome
	if sm.credStore != nil {
		cred, err := sm.credStore.Get(userID)
		if err == nil {
			log.Printf("Found credentials for %s, will auto-login", userID)
			session = NewSessionChromeWithCredentials(id, userID, port, debugPort, cred)
		} else {
			log.Printf("No credentials found for %s, manual login required", userID)
			session = NewSessionChrome(id, userID, port, debugPort)
		}
	} else {
		session = NewSessionChrome(id, userID, port, debugPort)
	}

	sm.mutex.Unlock()

	log.Printf("Creating session %s for user %s on port %d (debug port %d)", id, userID, port, debugPort)
	if session.credential != nil {
		log.Printf("Browser will open and auto-login with saved credentials...")
	} else {
		log.Printf("Browser will open - please complete Steam login...")
	}

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

// AddSessionWithTimeout creates a session with auto-removal after specified duration
func (sm *SessionManager) AddSessionWithTimeout(userID string, timeout time.Duration) (*SessionChrome, error) {
	session, err := sm.AddSession(userID)
	if err != nil {
		return nil, err
	}

	// Set up auto-removal timer
	sm.mutex.Lock()
	sm.timeouts[session.ID] = time.AfterFunc(timeout, func() {
		log.Printf("Session %s timeout reached, auto-removing...", session.ID)
		sm.RemoveSession(session.ID)
	})
	sm.mutex.Unlock()

	log.Printf("Session %s will auto-remove after %v", session.ID, timeout)

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

	// Cancel timeout if exists
	if timer, ok := sm.timeouts[id]; ok {
		timer.Stop()
		delete(sm.timeouts, id)
	}

	if err := session.Stop(); err != nil {
		return err
	}

	delete(sm.sessions, id)
	log.Printf("Session %s removed", id)

	return nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*SessionChrome, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []*SessionChrome {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make([]*SessionChrome, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// StopAll stops all sessions
func (sm *SessionManager) StopAll() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Stop all timers
	for _, timer := range sm.timeouts {
		timer.Stop()
	}
	sm.timeouts = make(map[string]*time.Timer)

	// Stop all sessions
	for id, session := range sm.sessions {
		if err := session.Stop(); err != nil {
			log.Printf("Error stopping session %s: %v", id, err)
		}
	}

	sm.sessions = make(map[string]*SessionChrome)
	log.Println("All sessions stopped")
}

// AddCredential adds Steam account credentials to the store
func (sm *SessionManager) AddCredential(cred *SteamCredential) error {
	if sm.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return sm.credStore.Add(cred)
}

// RemoveCredential removes credentials from the store
func (sm *SessionManager) RemoveCredential(accountName string) error {
	if sm.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return sm.credStore.Remove(accountName)
}

// ListCredentials returns all stored account names
func (sm *SessionManager) ListCredentials() []string {
	if sm.credStore == nil {
		return []string{}
	}
	return sm.credStore.List()
}

// ExtendTimeout extends the timeout for a session
func (sm *SessionManager) ExtendTimeout(id string, duration time.Duration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[id]; !exists {
		return fmt.Errorf("session %s not found", id)
	}

	// Cancel existing timer
	if timer, ok := sm.timeouts[id]; ok {
		timer.Stop()
	}

	// Set new timer
	sm.timeouts[id] = time.AfterFunc(duration, func() {
		log.Printf("Session %s timeout reached, auto-removing...", id)
		sm.RemoveSession(id)
	})

	log.Printf("Session %s timeout extended by %v", id, duration)

	return nil
}
