package steam

import (
	"context"
	"discordBot/util"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// SetProfileURLFromInput sets the ProfileURL field from a UID or a Steam profile URL.
// If input is a Steam64 UID or custom ID, constructs the profile URL. If input is a valid Steam profile URL, uses it directly.
func (s *Session) SetProfileURLFromInput(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		log.Printf("[SetProfileURLFromInput] Empty input, setting ProfileURL to empty string")
		s.ProfileURL = ""
		return
	}
	// If input looks like a Steam profile URL, use it directly
	if strings.HasPrefix(input, "https://steamcommunity.com/profiles/") || strings.HasPrefix(input, "https://steamcommunity.com/id/") {
		log.Printf("[SetProfileURLFromInput] Input is full profile URL: %s", input)
		s.ProfileURL = input
		return
	}
	// If input is a Steam64 ID
	if isSteam64(input) {
		log.Printf("[SetProfileURLFromInput] Input is Steam64 ID: %s", input)
		s.ProfileURL = "https://steamcommunity.com/profiles/" + input
		return
	}
	// Otherwise, treat as custom ID
	log.Printf("[SetProfileURLFromInput] Input is custom ID: %s", input)
	s.ProfileURL = "https://steamcommunity.com/id/" + input
}

// Session represents a logged-in Steam bot session with browser automation
type Session struct {
	ID     string
	UserID string
	Port   int
	// server      *http.Server // Removed: no HTTP server needed
	mutex         sync.Mutex
	isActive      bool
	ctx           context.Context
	cancel        context.CancelFunc
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	IsLoggedIn    bool
	loginChan     chan bool
	CreatedAt     time.Time
	credential    *util.SteamCredential
	debugPort     int
	profileDir    string
	chromeCmd     *exec.Cmd
	ProfileURL    string          // URL to open in new tab after login
	profileTabCtx context.Context // Store profile tab context for refresh
}

// InteractWithProfilePage performs the automated clicks on the profile page
func (s *Session) InteractWithProfilePage() {

	// First click
	err := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[1]/div[7]/div[6]/div[1]/div[2]/div/div/div/div[4]/div[2]/span/span`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[1]/div[7]/div[6]/div[1]/div[2]/div/div/div/div[4]/div[2]/span/span`, chromedp.BySearch),
	)
	if err != nil {
		log.Printf("[Session %s] Failed to click first element: %v", s.ID, err)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Second click (restored XPath)
	err2 := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[1]/div[7]/div[6]/div[1]/div[2]/div/div/div/div[4]/div[2]/div/div[9]/a[4]`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[1]/div[7]/div[6]/div[1]/div[2]/div/div/div/div[4]/div[2]/div/div[9]/a[4]`, chromedp.BySearch),
	)
	if err2 != nil {
		log.Printf("[Session %s] Failed to click second element (div[9]/a[4] XPath): %v", s.ID, err2)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	time.Sleep(1 * time.Second)
	// Third click (new XPath)
	err3 := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[6]`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[6]`, chromedp.BySearch),
	)
	if err3 != nil {
		log.Printf("[Session %s] Failed to click third element (div[6] XPath): %v", s.ID, err3)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Fourth click (new XPath)
	err4 := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[1]`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[1]`, chromedp.BySearch),
	)
	if err4 != nil {
		log.Printf("[Session %s] Failed to click fourth element (div[1] XPath): %v", s.ID, err4)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Fill out the textarea with a report reason
	reportReasons := []string{"cheating in cs2", "cheating in Counter-Strike 2", "using cheats in cs2", "suspected cheating in cs2"}
	for _, reason := range reportReasons {
		err := chromedp.Run(s.profileTabCtx,
			chromedp.WaitVisible(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/textarea`, chromedp.BySearch),
			chromedp.SetValue(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/textarea`, reason, chromedp.BySearch),
		)
		if err != nil {
			log.Printf("[Session %s] Failed to fill textarea with '%s': %v", s.ID, reason, err)
		} else {
			time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
			   // Stop after first successful fill
			   return
		}
	}

	// Click the next element after filling textarea
	err5 := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/div[1]/div[1]/div/div[1]`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/div[1]/div[1]/div/div[1]`, chromedp.BySearch),
	)
	if err5 != nil {
		log.Printf("[Session %s] Failed to click element after textarea: %v", s.ID, err5)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Click the final button
	err6 := chromedp.Run(s.profileTabCtx,
		chromedp.WaitVisible(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/button`, chromedp.BySearch),
		chromedp.Click(`/html/body/div[4]/div[3]/div/div/div/div[2]/div[2]/div[2]/button`, chromedp.BySearch),
	)
	if err6 != nil {
		log.Printf("[Session %s] Failed to click final button: %v", s.ID, err6)
	}
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

}

// NewSession creates a new bot session with ChromeDP
func NewSession(id, userID string, port, debugPort int) *Session {
	return &Session{
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

// NewSessionWithCredentials creates a new bot session with auto-login
func NewSessionWithCredentials(id, userID string, port, debugPort int, cred *util.SteamCredential) *Session {
	return &Session{
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

// isSteam64 checks if a string is a Steam64 ID (17 digits, all numeric)
func isSteam64(id string) bool {
	if len(id) != 17 {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// detectLoggedInState checks if we're actually logged in by looking for Steam UI elements
func (s *Session) detectLoggedInState() bool {
	// Check for the account name button by XPath and ensure it contains text
	var accountName string
	err := chromedp.Run(s.ctx,
		chromedp.Text(`/html/body/div[1]/div[7]/div[1]/div/div[3]/div/button`, &accountName, chromedp.BySearch, chromedp.NodeVisible),
	)
	if err == nil && strings.TrimSpace(accountName) != "" {
		return true
	}

	currentURL := s.GetCurrentURL()
	if strings.Contains(currentURL, "steamcommunity.com") ||
		(strings.Contains(currentURL, "store.steampowered.com") && !strings.Contains(currentURL, "/login")) {
		return true
	}

	return false
}

// performAutoLogin attempts to automatically log in using stored credentials
// performQRLogin is a stub for QR login automation. Implement as needed.
func (s *Session) performQRLogin() error {
	return fmt.Errorf("performQRLogin not implemented")
}

// performPasswordLogin is a stub for password login automation. Implement as needed.
func (s *Session) performPasswordLogin() error {
	if s.credential == nil {
		return fmt.Errorf("no credentials provided for password login")
	}
	log.Printf("[Session %s] Attempting password login for account: %s", s.ID, s.credential.AccountName)

	ctx := s.ctx
	if ctx == nil {
		return fmt.Errorf("chromedp context is nil in performPasswordLogin")
	}

	// Use provided absolute XPaths for reliability
	usernameXPath := `/html/body/div[1]/div[7]/div[7]/div[3]/div[1]/div/div/div/div[2]/div/form/div[1]/input`
	passwordXPath := `/html/body/div[1]/div[7]/div[7]/div[3]/div[1]/div/div/div/div[2]/div/form/div[2]/input`
	submitXPath := `/html/body/div[1]/div[7]/div[7]/div[3]/div[1]/div/div/div/div[2]/div/form/div[4]/button`

	// Wait for fields to be visible
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(usernameXPath, chromedp.BySearch),
		chromedp.WaitVisible(passwordXPath, chromedp.BySearch),
	)
	if err != nil {
		log.Printf("[Session %s] Login form fields not visible: %v", s.ID, err)
		return err
	}

	       // Focus and fill username and password, with delays
		       // Human-like delays
		       randDelay := func(minMs, maxMs int) { time.Sleep(time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond) }

		       err = chromedp.Run(ctx,
			       chromedp.Focus(usernameXPath, chromedp.BySearch),
		       )
		       randDelay(400, 900)
		       err = chromedp.Run(ctx,
			       chromedp.SetValue(usernameXPath, s.credential.AccountName, chromedp.BySearch),
		       )
		       if err != nil {
			       log.Printf("[Session %s] Failed to set username: %v", s.ID, err)
			       return err
		       }
		       randDelay(500, 1200)

		       err = chromedp.Run(ctx,
			       chromedp.Focus(passwordXPath, chromedp.BySearch),
		       )
		       randDelay(400, 900)
		       err = chromedp.Run(ctx,
			       chromedp.SetValue(passwordXPath, s.credential.Password, chromedp.BySearch),
		       )
		       if err != nil {
			       log.Printf("[Session %s] Failed to set password: %v", s.ID, err)
			       return err
		       }
		       randDelay(500, 1200)

		       // Wait for form to be stable before submitting
		       err = chromedp.Run(ctx,
			       chromedp.WaitVisible(submitXPath, chromedp.BySearch),
		       )
		       if err != nil {
			       log.Printf("[Session %s] Login button not visible: %v", s.ID, err)
			       return err
		       }
		       randDelay(300, 700)

		       // Re-set username field right before submit (if needed)
		       err = chromedp.Run(ctx,
			       chromedp.SetValue(usernameXPath, s.credential.AccountName, chromedp.BySearch),
		       )
		       randDelay(200, 500)

		       // Click the login button
		       err = chromedp.Run(ctx,
			       chromedp.Click(submitXPath, chromedp.BySearch),
		       )
		       if err != nil {
			       log.Printf("[Session %s] Failed to click login button: %v", s.ID, err)
			       return err
		       }

	log.Printf("[Session %s] Login form submitted, waiting for login to complete...", s.ID)
	return nil
}
func (s *Session) performAutoLogin() error {
	if s.credential == nil {
		return fmt.Errorf("no credentials provided")
	}

	// Handle QR code login differently
	if s.credential.LoginMethod == util.LoginMethodQR {
		return s.performQRLogin()
	}

	// Standard username/password login
	return s.performPasswordLogin()
}

// Start begins listening on the session's port and launches browser
func (s *Session) Start() error {
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

	// Wait for ChromeDP connection to be ready
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		if err := chromedp.Run(s.ctx, chromedp.Tasks{}); err == nil {
			break
		}
		if i == 19 {
			s.mutex.Unlock()
			return fmt.Errorf("could not connect to ChromeDP after retries")
		}
	}

	// Run login automation after navigating to login page
	if err := s.performAutoLogin(); err != nil {
		log.Printf("[Session %s] Login automation failed: %v", s.ID, err)
		return err
	}

	// Wait for login to complete (detect logged-in state)
	for i := 0; i < 15; i++ {
		if s.detectLoggedInState() {
			log.Printf("[Session %s] Login successful.", s.ID)
			break
		}
		time.Sleep(1 * time.Second)
		if i == 14 {
			return fmt.Errorf("login did not complete after retries")
		}
	}

	// Now navigate to the user-specified profile URL
	profileTabURL := s.ProfileURL
	if profileTabURL == "" {
		profileTabURL = "https://steamcommunity.com/"
	}
	s.profileTabCtx = s.ctx
	if err := chromedp.Run(s.profileTabCtx, chromedp.Navigate(profileTabURL)); err != nil {
		log.Printf("[Session %s] Failed to navigate to profile: %v", s.ID, err)
	}

	// After login, refresh the profile tab to transfer session
	if err := chromedp.Run(s.profileTabCtx, chromedp.Reload()); err != nil {
		log.Printf("[Session %s] Failed to refresh profile tab: %v", s.ID, err)
	} else {
		// Call the new profile interaction function
		s.InteractWithProfilePage()
	}

	// No HTTP server: just return after login
	s.mutex.Lock()
	s.isActive = true
	s.mutex.Unlock()
	return nil
}

// initBrowser launches Chrome with remote debugging and connects ChromeDP
func (s *Session) initBrowser() error {
	// ...existing code...
	// ...existing code...

	chromePath := "/usr/bin/google-chrome"
	s.profileDir = filepath.Join(".", "chrome-profiles", fmt.Sprintf("profile-%s", s.ID))

	// Create profile directory
	if err := os.MkdirAll(s.profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Clean up lock files
	os.Remove(filepath.Join(s.profileDir, "SingletonLock"))
	os.Remove(filepath.Join(s.profileDir, "SingletonSocket"))

	// Launch Chrome with chromedp's ExecAllocator
	loginURL := "https://store.steampowered.com/login/"
	ops := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.UserDataDir(s.profileDir),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1024,768"),
		chromedp.Flag("headless", false),
	)
	s.allocCtx, s.allocCancel = chromedp.NewExecAllocator(context.Background(), ops...)
	s.ctx, s.cancel = chromedp.NewContext(s.allocCtx)

	// Wait for browser to be ready
	for i := 0; i < 10; i++ {
		err := chromedp.Run(s.ctx, chromedp.Tasks{})
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Open login page in the first tab
	if err := chromedp.Run(s.ctx, chromedp.Navigate(loginURL)); err != nil {
		return fmt.Errorf("failed to open login tab: %w", err)
	}

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

	return nil
}

// navigateToSteamLogin navigates to Steam login page
func (s *Session) navigateToSteamLogin() error {
	loginURL := "https://store.steampowered.com/login/"
	if s.credential != nil && s.credential.LoginMethod == util.LoginMethodQR {
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
func (s *Session) NavigateTo(url string) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}
	return chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)
}


// GetCurrentURL returns the current URL of the session
func (s *Session) GetCurrentURL() string {
	var url string
	chromedp.Run(s.ctx,
		chromedp.Location(&url),
	)
	return url
}

// ClickElement finds and clicks an element by CSS selector
func (s *Session) ClickElement(selector string) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}

	return chromedp.Run(s.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// WaitForSelector waits for an element to appear
func (s *Session) WaitForSelector(selector string, timeoutMs int) error {
	if !s.IsLoggedIn {
		return fmt.Errorf("session not logged in")
	}

	ctx, cancel := context.WithTimeout(s.ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)
}

// monitorLogin waits for login, then navigates to bot profile and opens user-inputted profile in new tab
func (s *Session) monitorLogin() {
	for {
		var currentURL string
		err := chromedp.Run(s.ctx, chromedp.Location(&currentURL))
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if currentURL != "" &&
			!strings.Contains(currentURL, "store.steampowered.com/login") &&
			!strings.Contains(currentURL, "login.steampowered.com") &&
			!strings.Contains(currentURL, "help.steampowered.com/wizard/Login") {
			if s.detectLoggedInState() {
				log.Printf("[Session %s] Login successful! Staying on: %s", s.ID, currentURL)
				s.loginChan <- true
				break
			}
		}

	}
}
