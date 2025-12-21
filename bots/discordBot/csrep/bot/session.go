package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// Session represents a logged-in bot session with browser automation
type SessionChrome struct {
	ID          string
	UserID      string
	Port        int
	server      *http.Server
	mutex       sync.Mutex
	isActive    bool
	ctx         context.Context
	cancel      context.CancelFunc
	allocCtx    context.Context
	allocCancel context.CancelFunc
	IsLoggedIn  bool
	loginChan   chan bool
	CreatedAt   time.Time
	credential  *SteamCredential
	debugPort   int
	profileDir  string
	chromeCmd   *exec.Cmd
}

// NewSessionChrome creates a new bot session with ChromeDP
func NewSessionChrome(id, userID string, port, debugPort int) *SessionChrome {
	return &SessionChrome{
		ID:         id,
		UserID:     userID,
		Port:       port,
		debugPort:  debugPort,
		isActive:   false,
		IsLoggedIn: false,
		loginChan:  make(chan bool),
		CreatedAt:  time.Now(),
		credential: nil,
	}
}

// NewSessionChromeWithCredentials creates a new bot session with auto-login
func NewSessionChromeWithCredentials(id, userID string, port, debugPort int, cred *SteamCredential) *SessionChrome {
	return &SessionChrome{
		ID:         id,
		UserID:     userID,
		Port:       port,
		debugPort:  debugPort,
		isActive:   false,
		IsLoggedIn: false,
		loginChan:  make(chan bool),
		CreatedAt:  time.Now(),
		credential: cred,
	}
}

// Start begins listening on the session's port and launches browser
func (s *SessionChrome) Start() error {
	s.mutex.Lock()

	if s.isActive {
		s.mutex.Unlock()
		return fmt.Errorf("session %s already active", s.ID)
	}

	// Initialize Chrome and connect via ChromeDP
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

	// If credentials are provided, attempt auto-login
	if s.credential != nil {
		log.Printf("[Session %s] Auto-logging in with account: %s", s.ID, s.credential.AccountName)
		if err := s.performAutoLogin(); err != nil {
			log.Printf("[Session %s] Auto-login failed: %v, waiting for manual login", s.ID, err)
			log.Printf("[Session %s] Please log in manually in the browser window.", s.ID)
		}
	} else {
		log.Printf("[Session %s] Browser opened at Steam login. Waiting for login to complete...", s.ID)
		log.Printf("[Session %s] Please log in manually in the browser window.", s.ID)
	}

	// Start monitoring for login completion
	go s.monitorLogin()

	// Block until login is complete
	<-s.loginChan

	s.mutex.Lock()
	s.IsLoggedIn = true
	s.mutex.Unlock()

	log.Printf("[Session %s] ✓ Login detected! Starting API server...", s.ID)

	// Now start the HTTP server for automation
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
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
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[Session %s] Server error: %v", s.ID, err)
		}
	}()

	return nil
}

// initBrowser launches Chrome with remote debugging and connects ChromeDP
func (s *SessionChrome) initBrowser() error {
	chromePath := "/usr/bin/google-chrome"
	s.profileDir = filepath.Join(".", "chrome-profiles", fmt.Sprintf("profile-%s", s.ID))

	// Create profile directory
	if err := os.MkdirAll(s.profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Clean up lock files
	os.Remove(filepath.Join(s.profileDir, "SingletonLock"))
	os.Remove(filepath.Join(s.profileDir, "SingletonSocket"))

	// Launch Chrome with remote debugging
	args := []string{
		fmt.Sprintf("--user-data-dir=%s", s.profileDir),
		fmt.Sprintf("--remote-debugging-port=%d", s.debugPort),
		"--remote-debugging-address=127.0.0.1",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-blink-features=AutomationControlled",
		"--window-size=1024,768",
		"about:blank",
	}

	s.chromeCmd = exec.Command(chromePath, args...)
	if err := s.chromeCmd.Start(); err != nil {
		return fmt.Errorf("failed to launch Chrome: %w", err)
	}

	log.Printf("[Session %s] Chrome launched on debug port %d (PID: %d)", s.ID, s.debugPort, s.chromeCmd.Process.Pid)

	// Wait for Chrome to start
	time.Sleep(2 * time.Second)

	// Connect ChromeDP to the running Chrome instance
	s.allocCtx, s.allocCancel = chromedp.NewRemoteAllocator(context.Background(),
		fmt.Sprintf("http://127.0.0.1:%d", s.debugPort))

	s.ctx, s.cancel = chromedp.NewContext(s.allocCtx)

	// Test connection
	if err := chromedp.Run(s.ctx); err != nil {
		return fmt.Errorf("failed to connect to Chrome: %w", err)
	}

	return nil
}

// Stop shuts down the session server and closes browser
func (s *SessionChrome) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isActive {
		return nil
	}

	s.isActive = false

	// Close ChromeDP contexts
	if s.cancel != nil {
		s.cancel()
	}
	if s.allocCancel != nil {
		s.allocCancel()
	}

	// Kill Chrome process
	if s.chromeCmd != nil && s.chromeCmd.Process != nil {
		s.chromeCmd.Process.Kill()
	}

	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// navigateToSteamLogin navigates to Steam login page
func (s *SessionChrome) navigateToSteamLogin() error {
	loginURL := "https://store.steampowered.com/login/"
	if s.credential != nil && s.credential.LoginMethod == LoginMethodQR {
		loginURL = "https://steamcommunity.com/login/home/?goto="
		log.Printf("[Session %s] Navigating to Steam QR login page", s.ID)
	} else {
		log.Printf("[Session %s] Navigating to Steam login page", s.ID)
	}

	return chromedp.Run(s.ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitReady("body"),
	)
}

// NavigateTo navigates the session browser to a URL
func (s *SessionChrome) NavigateTo(url string) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}
	return chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)
}

// GetScreenshot returns a PNG screenshot of the current page
func (s *SessionChrome) GetScreenshot() ([]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var buf []byte
	err := chromedp.Run(s.ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	return buf, err
}

// GetCurrentURL returns the current URL of the session
func (s *SessionChrome) GetCurrentURL() string {
	var url string
	chromedp.Run(s.ctx,
		chromedp.Location(&url),
	)
	return url
}

// ClickElement finds and clicks an element by CSS selector
func (s *SessionChrome) ClickElement(selector string) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}

	return chromedp.Run(s.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// WaitForSelector waits for an element to appear
func (s *SessionChrome) WaitForSelector(selector string, timeoutMs int) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}

	ctx, cancel := context.WithTimeout(s.ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)
}

// monitorLogin continuously checks if the user has successfully logged into Steam
func (s *SessionChrome) monitorLogin() {
	for {
		var currentURL string
		err := chromedp.Run(s.ctx,
			chromedp.Location(&currentURL),
		)

		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Check if we've successfully navigated away from login page
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

		time.Sleep(500 * time.Millisecond)
	}
}

// detectLoggedInState checks if we're actually logged in by looking for Steam UI elements
func (s *SessionChrome) detectLoggedInState() bool {
	var exists bool
	err := chromedp.Run(s.ctx,
		chromedp.Evaluate(`!!document.querySelector('#account_pulldown, .account_menu, [class*="accountmenu"]')`, &exists),
	)

	if err == nil && exists {
		return true
	}

	currentURL := s.GetCurrentURL()
	if contains(currentURL, "steamcommunity.com") ||
		(contains(currentURL, "store.steampowered.com") && !contains(currentURL, "/login")) {
		return true
	}

	return false
}

// performAutoLogin attempts to automatically log in using stored credentials
func (s *SessionChrome) performAutoLogin() error {
	if s.credential == nil {
		return fmt.Errorf("no credentials provided")
	}

	// Handle QR code login differently
	if s.credential.LoginMethod == LoginMethodQR {
		return s.performQRLogin()
	}

	// Standard username/password login
	return s.performPasswordLogin()
}

// performPasswordLogin handles username + password authentication
func (s *SessionChrome) performPasswordLogin() error {
	// Wait for login form to load
	time.Sleep(2 * time.Second)

	return chromedp.Run(s.ctx,
		// Fill in username
		chromedp.WaitVisible(`input[type='text'], input[name='username']`, chromedp.ByQuery),
		chromedp.SendKeys(`input[type='text'], input[name='username']`, s.credential.AccountName, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		// Fill in password
		chromedp.WaitVisible(`input[type='password'], input[name='password']`, chromedp.ByQuery),
		chromedp.SendKeys(`input[type='password'], input[name='password']`, s.credential.Password, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		// Click sign in button
		chromedp.Click(`button[type='submit'], button.login_btn`, chromedp.ByQuery),
	)
}

// performQRLogin handles QR code based authentication
func (s *SessionChrome) performQRLogin() error {
	log.Printf("[Session %s] QR code login - waiting for user to scan QR code...", s.ID)
	log.Printf("[Session %s] Please scan the QR code with your Steam Mobile app", s.ID)
	// QR login is handled by user scanning the code
	return nil
}

// HTTP Handler functions

func (s *SessionChrome) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"sessionId": s.ID,
		"userId":    s.UserID,
		"port":      s.Port,
	})
}

func (s *SessionChrome) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":  s.ID,
		"userId":     s.UserID,
		"port":       s.Port,
		"active":     s.isActive,
		"isLoggedIn": s.IsLoggedIn,
		"currentUrl": s.GetCurrentURL(),
		"createdAt":  s.CreatedAt,
	})
}

func (s *SessionChrome) handleOpenProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.IsLoggedIn {
		http.Error(w, "Session not logged in yet", http.StatusPreconditionFailed)
		return
	}

	log.Printf("[Session %s] Executing action: open-profile for user %s", s.ID, s.UserID)

	url := fmt.Sprintf("https://steamcommunity.com/id/%s", s.UserID)
	err := s.NavigateTo(url)
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

func (s *SessionChrome) handleNavigate(w http.ResponseWriter, r *http.Request) {
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

	err := s.NavigateTo(payload.URL)
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

func (s *SessionChrome) handleGetScreenshot(w http.ResponseWriter, r *http.Request) {
	screenshot, err := s.GetScreenshot()
	if err != nil {
		http.Error(w, fmt.Sprintf("Screenshot failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(screenshot)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
