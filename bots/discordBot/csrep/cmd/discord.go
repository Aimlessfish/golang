package main

import (
	"csrep/bot"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ReportProcess tracks a single report with multiple bot sessions
type ReportProcess struct {
	ReportID   string
	TargetID   string
	ProfileURL string
	SessionIDs []string
	BotNames   []string
	StartTime  time.Time
	Status     string
}

// DiscordBot manages Discord integration with Steam sessions
type DiscordBot struct {
	manager       *bot.SessionManager
	token         string
	session       *discordgo.Session
	reports       map[string]*ReportProcess // reportID -> ReportProcess
	reportCounter int
	reportsMutex  sync.RWMutex
}

func NewDiscordBot(token string) *DiscordBot {
	return &DiscordBot{
		manager:       bot.NewSessionManager(startPort),
		token:         token,
		reports:       make(map[string]*ReportProcess),
		reportCounter: 1,
	}
}

func (d *DiscordBot) Start() error {
	if d.token == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN not set in .env file")
	}

	log.Printf("Starting Discord bot...")
	log.Printf("Session start port: %d", startPort)
	if defaultTimeout > 0 {
		log.Printf("Default session timeout: %d minutes", defaultTimeout)
	}

	// Create Discord session
	var err error
	d.session, err = discordgo.New("Bot " + d.token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	// Register message handler
	d.session.AddHandler(d.messageHandler)

	// Set intents
	d.session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Open connection
	err = d.session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}

	log.Println("✅ Discord bot is now running!")
	log.Println("\nSupported commands:")
	log.Println("  !bot-add <email> <username> <password> [shared_secret]  - Add bot account")
	log.Println("  !bot-list                                                - List bot accounts")
	log.Println("  !bot-remove <username>                                   - Remove bot account")
	log.Println("  !report <steam_profile_url>                              - Report a player")
	log.Println("  !sessions                                                - List active sessions")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down Discord bot...")
	d.session.Close()
	d.manager.StopAll()

	return nil
}

// cleanupMessages deletes previous messages in the DM channel to keep it clean
func (d *DiscordBot) cleanupMessages(s *discordgo.Session, channelID string, keepCount int) {
	// Fetch recent messages
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Printf("Error fetching messages for cleanup: %v", err)
		return
	}

	// Delete all but the last keepCount messages
	if len(messages) > keepCount {
		for i := keepCount; i < len(messages); i++ {
			err := s.ChannelMessageDelete(channelID, messages[i].ID)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
			time.Sleep(100 * time.Millisecond) // Rate limit protection
		}
	}
}

// messageHandler processes Discord messages
func (d *DiscordBot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot's own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Only respond to DMs (private channels) - ignore server messages
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Try fetching from API if not in state
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			log.Printf("Error getting channel: %v", err)
			return
		}
	}

	// Only respond in DMs
	if channel.Type != discordgo.ChannelTypeDM {
		// If someone tries to use bot in a server channel, inform them
		if strings.HasPrefix(m.Content, "!") {
			s.ChannelMessageSend(m.ChannelID, "⚠️ Please DM me directly to use bot commands for privacy and security.")
		}
		return
	}

	// Only respond to messages starting with !
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// Parse command
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0][1:] // Remove !
	args := parts[1:]

	switch cmd {
	case "bot-add":
		d.handleBotAdd(s, m, args)
	case "bot-list":
		d.handleBotList(s, m)
	case "bot-remove":
		d.handleBotRemove(s, m, args)
	case "report":
		d.handleReport(s, m, args)
	case "reports":
		d.handleReports(s, m)
	case "help":
		d.handleHelp(s, m)
	}
}

// handleBotAdd adds a new bot account to the credential pool
func (d *DiscordBot) handleBotAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, ".\n Usage: `!bot-add <email> <username> <password> [shared_secret]`")
		return
	}

	email := args[0]
	username := args[1]
	sharedSecret := ""
	if len(args) >= 4 {
		sharedSecret = args[3]
	}

	// Mock response for testing
	s.ChannelMessageDelete(m.ChannelID, m.ID)

	response := fmt.Sprintf("✅ Bot account added: **%s**\n", username)
	response += fmt.Sprintf("📧 Email: %s\n", email)
	if sharedSecret != "" {
		response += "🔐 2FA: Enabled\n"
	} else {
		response += "🔐 2FA: Disabled\n"
	}
	response += "\n_Original message deleted for security_"

	s.ChannelMessageSend(m.ChannelID, response)

	// Cleanup old messages, keep only last 5
	go d.cleanupMessages(s, m.ChannelID, 5)
}

// handleBotList lists all stored bot accounts
func (d *DiscordBot) handleBotList(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Mock response for testing
	mockAccounts := []string{
		"test_bot_1",
		"test_bot_2",
		"test_bot_3",
	}

	response := "**Available Bot Accounts:**\n```\n"
	for i, acc := range mockAccounts {
		response += fmt.Sprintf("%d. %s\n", i+1, acc)
	}
	response += "```"

	s.ChannelMessageSend(m.ChannelID, response)

	// Cleanup old messages, keep only last 5
	go d.cleanupMessages(s, m.ChannelID, 5)
}

// handleBotRemove removes a bot account from the pool
func (d *DiscordBot) handleBotRemove(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "❌ Usage: `!bot-remove <username>`")
		return
	}

	username := args[0]
	// Mock response for testing
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Bot account removed: **%s**", username))

	// Cleanup old messages, keep only last 5
	go d.cleanupMessages(s, m.ChannelID, 5)
}

// handleReport reports a player using an available bot account
func (d *DiscordBot) handleReport(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!report <steam_profile_url> [number_of_bots]`\nExample: `!report https://steamcommunity.com/profiles/123456 5`")
		return
	}

	profileURL := args[0]

	// Parse number of bots (default: 1)
	numBots := 1
	if len(args) >= 2 {
		fmt.Sscanf(args[1], "%d", &numBots)
		if numBots < 1 {
			numBots = 1
		}
	}

	// Step 1: Validate URL and extract Steam ID
	steamID, err := extractSteamID(profileURL)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Invalid Steam profile URL: %v", err))
		return
	}

	// Mock response for testing
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Valid Steam profile detected\n🎯 Target Steam ID: `%s`\n🤖 Deploying **%d** bot(s)", steamID, numBots))

	// Mock bot launch results
	reportURL := fmt.Sprintf("https://help.steampowered.com/en/wizard/HelpWithGameIssue/?appid=730&issueid=1010&playerid=%s", steamID)
	var sessionIDs []string
	var botNames []string
	var successList []string

	for i := 0; i < numBots; i++ {
		botName := fmt.Sprintf("test_bot_%d", i+1)
		sessionID := fmt.Sprintf("session-%d", i+1)
		port := 9222 + i
		sessionIDs = append(sessionIDs, sessionID)
		botNames = append(botNames, botName)
		successList = append(successList, fmt.Sprintf("• **%s** → Session `%s` (port %d)", botName, sessionID, port))
	}

	summary := fmt.Sprintf("\n**📊 Report Summary:**\n✅ Successful: %d/%d bots\n", numBots, numBots)
	summary += "\n**Active Bots:**\n"
	for _, bot := range successList {
		summary += bot + "\n"
	}
	summary += fmt.Sprintf("\n🔗 Report URL: `%s`\n", reportURL)
	summary += "⏳ Ready for automation\n\n_Next: Add automation to select reason and submit_"

	s.ChannelMessageSend(m.ChannelID, summary)

	// Track this report process
	d.reportsMutex.Lock()
	reportID := fmt.Sprintf("report-%d", d.reportCounter)
	d.reportCounter++

	report := &ReportProcess{
		ReportID:   reportID,
		TargetID:   steamID,
		ProfileURL: profileURL,
		SessionIDs: sessionIDs,
		BotNames:   botNames,
		StartTime:  time.Now(),
		Status:     "active (mock)",
	}

	d.reports[reportID] = report
	d.reportsMutex.Unlock()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\n📋 Report tracked as: `%s`", reportID))

	// Cleanup old messages, keep only last 10 for report tracking
	go d.cleanupMessages(s, m.ChannelID, 10)

	// TODO: Add automated report submission for all sessions
}

// extractSteamID extracts the Steam ID from various Steam profile URL formats
func extractSteamID(url string) (string, error) {
	// Handle different URL formats:
	// https://steamcommunity.com/profiles/76561198012345678
	// https://steamcommunity.com/id/username
	// https://steamcommunity.com/profiles/[U:1:12345678]

	if strings.Contains(url, "/profiles/") {
		// Extract numeric Steam ID or [U:1:X] format
		parts := strings.Split(url, "/profiles/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid profile URL format")
		}
		steamID := strings.TrimSuffix(parts[1], "/")
		steamID = strings.Split(steamID, "?")[0] // Remove query params
		return steamID, nil
	} else if strings.Contains(url, "/id/") {
		// For vanity URLs, we need the custom ID
		// Steam's report page may accept vanity URLs or you may need to resolve them
		parts := strings.Split(url, "/id/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid custom URL format")
		}
		customID := strings.TrimSuffix(parts[1], "/")
		customID = strings.Split(customID, "?")[0] // Remove query params

		// Note: For vanity URLs, you might need to resolve to Steam ID64
		// For now, we'll try using the vanity URL directly
		// If that doesn't work, you'll need to call Steam API to resolve it
		return customID, nil
	}

	return "", fmt.Errorf("unrecognized Steam profile URL format")
}

// handleReports lists active report processes
func (d *DiscordBot) handleReports(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.reportsMutex.RLock()
	defer d.reportsMutex.RUnlock()

	if len(d.reports) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No active reports")
		return
	}

	response := "**📋 Active Report Processes:**\n\n"

	for _, report := range d.reports {
		elapsed := time.Since(report.StartTime).Round(time.Second)
		response += fmt.Sprintf("**%s**\n", report.ReportID)
		response += fmt.Sprintf("🎯 Target: `%s`\n", report.TargetID)
		response += fmt.Sprintf("🔗 URL: %s\n", report.ProfileURL)
		response += fmt.Sprintf("🤖 Bots: %d active\n", len(report.BotNames))
		response += fmt.Sprintf("⏱️ Running: %s\n", elapsed)
		response += fmt.Sprintf("🟢 Status: %s\n", report.Status)
		response += "```\n"
		for i, botName := range report.BotNames {
			response += fmt.Sprintf("%d. %s → %s\n", i+1, botName, report.SessionIDs[i])
		}
		response += "```\n\n"
	}

	s.ChannelMessageSend(m.ChannelID, response)

	// Cleanup old messages, keep only last 5
	go d.cleanupMessages(s, m.ChannelID, 5)
}

// handleHelp shows available commands
func (d *DiscordBot) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	help := `**🤖 Steam Bot Commands**

**Bot Management:**
` + "`!bot-add <email> <username> <password> [shared_secret]`" + ` - Add bot account
` +
		"`!bot-list`" + ` - List all bot accounts
` +
		"`!bot-remove <username>`" + ` - Remove bot account

` +
		"**Actions:**\n" +
		"`!report <steam_profile_url>`" + ` - Report a player (automated)
` +
		"`!sessions`" + ` - View active browser sessions

` +
		"**Info:**\n" +
		"`!help`" + ` - Show this help message

` +
		"_All automation is handled automatically - just add bot accounts and use commands!_"

	s.ChannelMessageSend(m.ChannelID, help)

	// Cleanup old messages, keep only last 3
	go d.cleanupMessages(s, m.ChannelID, 3)
}

// Command handlers for Discord

func (d *DiscordBot) handleAddSession(userID string, timeout int) (string, error) {
	var session *bot.SessionChrome
	var err error

	if timeout > 0 {
		duration := time.Duration(timeout) * time.Minute
		session, err = d.manager.AddSessionWithTimeout(userID, duration)
	} else {
		session, err = d.manager.AddSession(userID)
	}

	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("✅ Session '%s' created for %s on port %d\n", session.ID, userID, session.Port)
	if timeout > 0 {
		response += fmt.Sprintf("Auto-remove after %d minutes\n", timeout)
	}
	response += fmt.Sprintf("API: http://localhost:%d", session.Port)

	return response, nil
}

func (d *DiscordBot) handleRemoveSession(sessionID string) (string, error) {
	err := d.manager.RemoveSession(sessionID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ Session '%s' removed", sessionID), nil
}

func (d *DiscordBot) handleListSessions() string {
	sessions := d.manager.ListSessions()
	if len(sessions) == 0 {
		return "No active sessions"
	}

	response := "**Active Sessions:**\n```\n"
	response += fmt.Sprintf("%-12s %-15s %-6s %-10s\n", "ID", "Username", "Port", "Status")
	response += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	for _, s := range sessions {
		status := "Ready"
		if !s.IsLoggedIn {
			status = "Logging in"
		}
		response += fmt.Sprintf("%-12s %-15s %-6d %-10s\n", s.ID, s.UserID, s.Port, status)
	}

	response += "```"
	return response
}

func (d *DiscordBot) handleScreenshot(sessionID string) ([]byte, error) {
	session, err := d.manager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.GetScreenshot()
}

func (d *DiscordBot) handleNavigate(sessionID, url string) (string, error) {
	session, err := d.manager.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	if err := session.NavigateTo(url); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Session '%s' navigated to: %s", sessionID, url), nil
}
