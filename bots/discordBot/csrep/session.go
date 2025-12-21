package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Session represents a logged-in bot session with browser automation
type Session struct {
	ID         string
	UserID     string
	Port       int
	server     *http.Server
	mutex      sync.Mutex
	isActive   bool
	browser    playwright.Browser
	context    playwright.BrowserContext
	page       playwright.Page
	pw         *playwright.Playwright
	isLoggedIn bool
	loginChan  chan bool
}

// NewSession creates a new bot session
func NewSession(id, userID string, port int) *Session {
	return &Session{
		ID:         id,
		UserID:     userID,
		Port:       port,
		isActive:   false,
		isLoggedIn: false,
		loginChan:  make(chan bool),
	}
}

// Start begins listening on the session's port and launches browser
// This method blocks until the user completes the Steam login
func (s *Session) Start() error {
	s.mutex.Lock()

	if s.isActive {
		s.mutex.Unlock()
		return fmt.Errorf("session %s already active", s.ID)
	}

	// Initialize Playwright and browser
	if err := s.initBrowser(); err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to initialize browser: %w", err)
	}

	// Navigate to Steam login page
	if err := s.navigateToSteamLogin(); err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to navigate to Steam login: %w", err)
	}

	s.mutex.Unlock()

	log.Printf("[Session %s] Browser opened at Steam login. Waiting for login to complete...", s.ID)
	log.Printf("[Session %s] Please log in manually in the browser window.", s.ID)

	// Start monitoring for login completion
	go s.monitorLogin()

	// Block until login is complete
	<-s.loginChan

	s.mutex.Lock()
	s.isLoggedIn = true
	s.mutex.Unlock()

	log.Printf("[Session %s] ✓ Login detected! Starting API server...", s.ID)

	// Now start the HTTP server for automation
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Action trigger endpoints
	mux.HandleFunc("/action/open-profile", s.handleOpenProfile)
	mux.HandleFunc("/action/navigate", s.handleNavigate)
	mux.HandleFunc("/action/get-screenshot", s.handleGetScreenshot)
	mux.HandleFunc("/status", s.handleStatus)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	s.mutex.Lock()
	s.isActive = true
	s.mutex.Unlock()

	go func() {
		log.Printf("[Session %s] API server ready on port %d", s.ID, s.Port)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[Session %s] Server error: %v", s.ID, err)
		}
	}()

	return nil
}

// Stop shuts down the session server and closes browser
func (s *Session) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isActive {
		return nil
	}

	s.isActive = false

	// Close browser resources
	if s.page != nil {
		s.page.Close()
	}
	if s.context != nil {
		s.context.Close()
	}
	if s.browser != nil {
		s.browser.Close()
	}
	if s.pw != nil {
		s.pw.Stop()
	}

	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Handler functions

func (s *Session) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"sessionId": s.ID,
		"userId":    s.UserID,
		"port":      s.Port,
	})
}

func (s *Session) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":  s.ID,
		"userId":     s.UserID,
		"port":       s.Port,
		"active":     s.isActive,
		"isLoggedIn": s.isLoggedIn,
		"currentUrl": s.getCurrentURL(),
	})
}

func (s *Session) handleOpenProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isLoggedIn {
		http.Error(w, "Session not logged in yet", http.StatusPreconditionFailed)
		return
	}

	log.Printf("[Session %s] Executing action: open-profile for user %s", s.ID, s.UserID)

	url := fmt.Sprintf("https://steamcommunity.com/id/%s", s.UserID)
	err := s.navigateTo(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"action":  "open-profile",
		"userId":  s.UserID,
		"message": fmt.Sprintf("Opened profile for %s", s.UserID),
	})
}

func (s *Session) handleNavigate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	log.Printf("[Session %s] Executing action: navigate to %s", s.ID, payload.URL)

	err := s.navigateTo(payload.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"action":  "navigate",
		"url":     payload.URL,
		"message": fmt.Sprintf("Navigated to %s", payload.URL),
	})
}

func (s *Session) handleGetScreenshot(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.page == nil {
		http.Error(w, "Browser not initialized", http.StatusInternalServerError)
		return
	}

	screenshot, err := s.page.Screenshot()
	if err != nil {
		http.Error(w, fmt.Sprintf("Screenshot failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(screenshot)
}

// Helper methods

func (s *Session) initBrowser() error {
	var err error
	s.pw, err = playwright.Run()
	if err != nil {
		return err
	}

	s.browser, err = s.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // Visible browser for manual login
		Args: []string{
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		return err
	}

	s.context, err = s.browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return err
	}

	s.page, err = s.context.NewPage()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) navigateToSteamLogin() error {
	if s.page == nil {
		return fmt.Errorf("browser not initialized")
	}

	log.Printf("[Session %s] Navigating to Steam login page", s.ID)
	_, err := s.page.Goto("https://store.steampowered.com/login/")
	return err
}

func (s *Session) navigateTo(url string) error {
	if s.page == nil {
		return fmt.Errorf("browser not initialized")
	}

	_, err := s.page.Goto(url)
	return err
}

func (s *Session) getCurrentURL() string {
	if s.page == nil {
		return ""
	}
	return s.page.URL()
}

// monitorLogin continuously checks if the user has successfully logged into Steam
func (s *Session) monitorLogin() {
	for {
		if s.page == nil {
			return
		}

		currentURL := s.page.URL()

		// Check if we've successfully navigated away from login page
		// Steam redirects to store or community after successful login
		if currentURL != "" &&
			!contains(currentURL, "store.steampowered.com/login") &&
			!contains(currentURL, "login.steampowered.com") &&
			!contains(currentURL, "help.steampowered.com/wizard/Login") {

			// Additional check: look for Steam-specific logged-in elements
			if s.detectLoggedInState() {
				log.Printf("[Session %s] Login successful! Redirected to: %s", s.ID, currentURL)
				s.loginChan <- true
				return
			}
		}

		// Check every 500ms
		time.Sleep(500 * time.Millisecond)
	}
}

// detectLoggedInState checks if we're actually logged in by looking for Steam UI elements
func (s *Session) detectLoggedInState() bool {
	if s.page == nil {
		return false
	}

	// Check for common logged-in indicators
	// Steam shows user account name when logged in
	account, err := s.page.QuerySelector("#account_pulldown, .account_menu, [class*='accountmenu']")
	if err == nil && account != nil {
		return true
	}

	// Check if we're on a Steam page that requires login
	currentURL := s.page.URL()
	if contains(currentURL, "steamcommunity.com") ||
		(contains(currentURL, "store.steampowered.com") && !contains(currentURL, "/login")) {
		return true
	}

	return false
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func findSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
