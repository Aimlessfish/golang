package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// LoginMethod defines how the session should authenticate
type LoginMethod string

const (
	LoginMethodPassword LoginMethod = "password" // Username + Password
	LoginMethodQR       LoginMethod = "qr"       // QR code login
)

// SteamCredential stores Steam account login information
type SteamCredential struct {
	AccountName    string      `json:"account_name"`
	Email          string      `json:"email,omitempty"` // Email for bot account
	Password       string      `json:"password,omitempty"`
	SharedSecret   string      `json:"shared_secret,omitempty"` // For Steam Guard 2FA
	IdentitySecret string      `json:"identity_secret,omitempty"`
	LoginMethod    LoginMethod `json:"login_method"` // "password" or "qr"
}

// CredentialStore manages Steam account credentials
type CredentialStore struct {
	credentials map[string]*SteamCredential // key: account_name
	mutex       sync.RWMutex
	filePath    string
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(filePath string) (*CredentialStore, error) {
	store := &CredentialStore{
		credentials: make(map[string]*SteamCredential),
		filePath:    filePath,
	}

	// Load existing credentials if file exists
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// Add adds or updates a credential
func (cs *CredentialStore) Add(cred *SteamCredential) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.credentials[cred.AccountName] = cred
	return cs.save()
}

// Get retrieves a credential by account name
func (cs *CredentialStore) Get(accountName string) (*SteamCredential, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	cred, exists := cs.credentials[accountName]
	if !exists {
		return nil, fmt.Errorf("credential not found for account: %s", accountName)
	}

	return cred, nil
}

// Remove removes a credential
func (cs *CredentialStore) Remove(accountName string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.credentials, accountName)
	return cs.save()
}

// List returns all account names
func (cs *CredentialStore) List() []string {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	accounts := make([]string, 0, len(cs.credentials))
	for name := range cs.credentials {
		accounts = append(accounts, name)
	}
	return accounts
}

// load reads credentials from file
func (cs *CredentialStore) load() error {
	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cs.credentials)
}

// save writes credentials to file
func (cs *CredentialStore) save() error {
	data, err := json.MarshalIndent(cs.credentials, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cs.filePath, data, 0600)
}
