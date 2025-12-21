package module

import (
	"csrep/bot"
	"fmt"
	"log"
	"strings"
	"sync"
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

// SteamReportModule is a Discord bot module for Steam reporting
type SteamReportModule struct {
	manager       *bot.SessionManager
	reports       map[string]*ReportProcess
	reportCounter int
	reportsMutex  sync.RWMutex
	startPort     int
	commandPrefix string // e.g., "steam-" or "report-"
}

// NewSteamReportModule creates a new Steam report module
func NewSteamReportModule(startPort int, commandPrefix string) *SteamReportModule {
	return &SteamReportModule{
		manager:       bot.NewSessionManager(startPort),
		reports:       make(map[string]*ReportProcess),
		reportCounter: 1,
		startPort:     startPort,
		commandPrefix: commandPrefix,
	}
}

// RegisterHandlers registers the module's message handlers with an existing Discord session
func (m *SteamReportModule) RegisterHandlers(s *discordgo.Session) {
	s.AddHandler(m.messageHandler)
	log.Printf("Steam Report Module registered with prefix: %s", m.commandPrefix)
}

// GetCommands returns a list of commands this module provides
func (m *SteamReportModule) GetCommands() []string {
	return []string{
		m.commandPrefix + "bot-add",
		m.commandPrefix + "bot-list",
		m.commandPrefix + "bot-remove",
		m.commandPrefix + "report",
		m.commandPrefix + "reports",
		m.commandPrefix + "help",
	}
}

// Shutdown gracefully shuts down the module
func (m *SteamReportModule) Shutdown() {
	log.Println("Shutting down Steam Report Module...")
	m.manager.StopAll()
}

// messageHandler processes Discord messages for this module
func (m *SteamReportModule) messageHandler(s *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore bot's own messages
	if msg.Author.ID == s.State.User.ID {
		return
	}

	// Only respond to DMs
	channel, err := s.State.Channel(msg.ChannelID)
	if err != nil {
		channel, err = s.Channel(msg.ChannelID)
		if err != nil {
			return
		}
	}

	if channel.Type != discordgo.ChannelTypeDM {
		return
	}

	// Check if message starts with our command prefix
	if !strings.HasPrefix(msg.Content, "!"+m.commandPrefix) {
		return
	}

	// Parse command
	content := strings.TrimPrefix(msg.Content, "!")
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}

	cmd := strings.TrimPrefix(parts[0], m.commandPrefix)
	args := parts[1:]

	switch cmd {
	case "bot-add":
		m.handleBotAdd(s, msg, args)
	case "bot-list":
		m.handleBotList(s, msg)
	case "bot-remove":
		m.handleBotRemove(s, msg, args)
	case "report":
		m.handleReport(s, msg, args)
	case "reports":
		m.handleReports(s, msg)
	case "help":
		m.handleHelp(s, msg)
	}
}

// cleanupMessages deletes previous messages in the DM channel to keep it clean
func (m *SteamReportModule) cleanupMessages(s *discordgo.Session, channelID string, keepCount int) {
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Printf("Error fetching messages for cleanup: %v", err)
		return
	}

	if len(messages) > keepCount {
		for i := keepCount; i < len(messages); i++ {
			err := s.ChannelMessageDelete(channelID, messages[i].ID)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// handleBotAdd adds a new bot account to the credential pool
func (m *SteamReportModule) handleBotAdd(s *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(msg.ChannelID, "❌ Usage: `!"+m.commandPrefix+"bot-add <email> <username> <password> [shared_secret]`")
		return
	}

	email := args[0]
	username := args[1]
	sharedSecret := ""
	if len(args) >= 4 {
		sharedSecret = args[3]
	}

	s.ChannelMessageDelete(msg.ChannelID, msg.ID)

	response := fmt.Sprintf("✅ Bot account added: **%s**\n", username)
	response += fmt.Sprintf("📧 Email: %s\n", email)
	if sharedSecret != "" {
		response += "🔐 2FA: Enabled\n"
	} else {
		response += "🔐 2FA: Disabled\n"
	}
	response += "\n_Original message deleted for security_"

	s.ChannelMessageSend(msg.ChannelID, response)
	go m.cleanupMessages(s, msg.ChannelID, 5)
}

// handleBotList lists all stored bot accounts
func (m *SteamReportModule) handleBotList(s *discordgo.Session, msg *discordgo.MessageCreate) {
	mockAccounts := []string{
		"test_bot_1",
		"test_bot_2",
		"test_bot_3",
	}

	response := "**🤖 Available Bot Accounts:**\n```\n"
	for i, acc := range mockAccounts {
		response += fmt.Sprintf("%d. %s\n", i+1, acc)
	}
	response += "```"

	s.ChannelMessageSend(msg.ChannelID, response)
	go m.cleanupMessages(s, msg.ChannelID, 5)
}

// handleBotRemove removes a bot account from the pool
func (m *SteamReportModule) handleBotRemove(s *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(msg.ChannelID, "❌ Usage: `!"+m.commandPrefix+"bot-remove <username>`")
		return
	}

	username := args[0]
	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("✅ Bot account removed: **%s**", username))
	go m.cleanupMessages(s, msg.ChannelID, 5)
}

// handleReport reports a player using available bot accounts
func (m *SteamReportModule) handleReport(s *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(msg.ChannelID, "❌ Usage: `!"+m.commandPrefix+"report <steam_profile_url> [number_of_bots]`\nExample: `!"+m.commandPrefix+"report https://steamcommunity.com/profiles/123456 5`")
		return
	}

	profileURL := args[0]
	numBots := 1
	if len(args) >= 2 {
		fmt.Sscanf(args[1], "%d", &numBots)
		if numBots < 1 {
			numBots = 1
		}
	}

	steamID, err := extractSteamID(profileURL)
	if err != nil {
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("❌ Invalid Steam profile URL: %v", err))
		return
	}

	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("✅ Valid Steam profile detected\n🎯 Target Steam ID: `%s`\n🤖 Deploying **%d** bot(s)", steamID, numBots))

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

	s.ChannelMessageSend(msg.ChannelID, summary)

	m.reportsMutex.Lock()
	reportID := fmt.Sprintf("report-%d", m.reportCounter)
	m.reportCounter++

	report := &ReportProcess{
		ReportID:   reportID,
		TargetID:   steamID,
		ProfileURL: profileURL,
		SessionIDs: sessionIDs,
		BotNames:   botNames,
		StartTime:  time.Now(),
		Status:     "active (mock)",
	}

	m.reports[reportID] = report
	m.reportsMutex.Unlock()

	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("\n📋 Report tracked as: `%s`", reportID))
	go m.cleanupMessages(s, msg.ChannelID, 10)
}

// extractSteamID extracts the Steam ID from various Steam profile URL formats
func extractSteamID(url string) (string, error) {
	if strings.Contains(url, "/profiles/") {
		parts := strings.Split(url, "/profiles/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid profile URL format")
		}
		steamID := strings.TrimSuffix(parts[1], "/")
		steamID = strings.Split(steamID, "?")[0]
		return steamID, nil
	} else if strings.Contains(url, "/id/") {
		parts := strings.Split(url, "/id/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid custom URL format")
		}
		customID := strings.TrimSuffix(parts[1], "/")
		customID = strings.Split(customID, "?")[0]
		return customID, nil
	}

	return "", fmt.Errorf("unrecognized Steam profile URL format")
}

// handleReports lists active report processes
func (m *SteamReportModule) handleReports(s *discordgo.Session, msg *discordgo.MessageCreate) {
	m.reportsMutex.RLock()
	defer m.reportsMutex.RUnlock()

	if len(m.reports) == 0 {
		s.ChannelMessageSend(msg.ChannelID, "No active reports")
		return
	}

	response := "**📋 Active Report Processes:**\n\n"

	for _, report := range m.reports {
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

	s.ChannelMessageSend(msg.ChannelID, response)
	go m.cleanupMessages(s, msg.ChannelID, 5)
}

// handleHelp shows available commands
func (m *SteamReportModule) handleHelp(s *discordgo.Session, msg *discordgo.MessageCreate) {
	help := fmt.Sprintf(`**🤖 Steam Report Module Commands**

**Bot Management:**
`+"`!%sbot-add <email> <username> <password> [shared_secret]`"+` - Add bot account
`+"`!%sbot-list`"+` - List all bot accounts
`+"`!%sbot-remove <username>`"+` - Remove bot account

**Actions:**
`+"`!%sreport <steam_profile_url> [num_bots]`"+` - Report a player
`+"`!%sreports`"+` - View active reports
`+"`!%shelp`"+` - Show this help message

_All commands work in DMs only for privacy_`,
		m.commandPrefix, m.commandPrefix, m.commandPrefix,
		m.commandPrefix, m.commandPrefix, m.commandPrefix)

	s.ChannelMessageSend(msg.ChannelID, help)
	go m.cleanupMessages(s, msg.ChannelID, 3)
}
