package steam

import (
	"discordBot/util"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var manager *SessionManager

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

var (
	reports       = make(map[string]*ReportProcess)
	reportCounter = 1
	reportsMutex  sync.RWMutex
)

// InitSteamManager initializes the global Steam session manager
func InitSteamManager(startPort int) {
	manager = NewSessionManager(startPort)
}

// HandleSteamCommands processes Steam-related Discord commands
func HandleSteamCommands(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// Only respond to DMs
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			return false
		}
	}

	if channel.Type != discordgo.ChannelTypeDM {
		if strings.HasPrefix(m.Content, "!steam") {
			s.ChannelMessageSend(m.ChannelID, "Please DM me directly to use bot commands for privacy and security.")
		}
		return false
	}

	content := strings.ToLower(m.Content)

	switch {
	case strings.HasPrefix(content, "!steam bot-add"):
		handleBotAdd(s, m)
		return true
	case strings.HasPrefix(content, "!steam bot-list"):
		handleBotList(s, m)
		return true
	case strings.HasPrefix(content, "!steam bot-remove"):
		handleBotRemove(s, m)
		return true
	case strings.HasPrefix(content, "!steam report "), strings.HasPrefix(content, "!report "):
		handleReport(s, m)
		return true
	case strings.HasPrefix(content, "!steam reports"), content == "!reports":
		handleReports(s, m)
		return true
	case content == "!steam" || content == "!steam help":
		handleSteamHelp(s, m)
		return true
	}

	return false
}

// cleanupMessages deletes previous messages in the DM channel to keep it clean
func cleanupMessages(s *discordgo.Session, channelID string, keepCount int) {
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

func handleBotAdd(s *discordgo.Session, m *discordgo.MessageCreate) {
	if manager == nil {
		s.ChannelMessageSend(m.ChannelID, "Steam manager not initialized")
		return
	}

	parts := strings.Fields(m.Content)
	if len(parts) < 5 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!steam bot-add <email> <username> <password> [shared_secret]`")
		return
	}

	email := parts[2]
	username := parts[3]
	password := parts[4]
	sharedSecret := ""
	if len(parts) >= 6 {
		sharedSecret = parts[5]
	}

	// Create credential
	cred := &util.SteamCredential{
		AccountName:    username,
		Email:          email,
		Password:       password,
		SharedSecret:   sharedSecret,
		IdentitySecret: "",
		LoginMethod:    util.LoginMethodPassword,
	}

	// Add to credential store
	err := manager.AddCredential(cred)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error adding bot account: %v", err))
		return
	}

	// Delete the original message for security
	s.ChannelMessageDelete(m.ChannelID, m.ID)

	response := fmt.Sprintf("Bot account added: **%s**\n", username)
	response += fmt.Sprintf("Email: %s\n", email)
	if sharedSecret != "" {
		response += "2FA: Enabled\n"
	} else {
		response += "2FA: Disabled\n"
	}
	response += "\n_Original message deleted for security_"

	s.ChannelMessageSend(m.ChannelID, response)
	go cleanupMessages(s, m.ChannelID, 5)
}

func handleBotList(s *discordgo.Session, m *discordgo.MessageCreate) {
	if manager == nil {
		s.ChannelMessageSend(m.ChannelID, "Steam manager not initialized")
		return
	}

	accounts := manager.ListCredentials()
	if len(accounts) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No bot accounts stored")
		return
	}

	response := "**Available Bot Accounts:**\n```\n"
	for i, acc := range accounts {
		response += fmt.Sprintf("%d. %s\n", i+1, acc)
	}
	response += "```"

	s.ChannelMessageSend(m.ChannelID, response)
	go cleanupMessages(s, m.ChannelID, 5)
}

func handleBotRemove(s *discordgo.Session, m *discordgo.MessageCreate) {
	if manager == nil {
		s.ChannelMessageSend(m.ChannelID, "Steam manager not initialized")
		return
	}

	parts := strings.Fields(m.Content)
	if len(parts) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!steam bot-remove <username>`")
		return
	}

	username := parts[2]
	err := manager.RemoveCredential(username)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Bot account removed: **%s**", username))
	go cleanupMessages(s, m.ChannelID, 5)
}

func handleReport(s *discordgo.Session, m *discordgo.MessageCreate) {
	if manager == nil {
		s.ChannelMessageSend(m.ChannelID, "Steam manager not initialized")
		return
	}

	parts := strings.Fields(m.Content)

	// Determine if using !report or !steam report format
	urlIndex := 1
	if strings.HasPrefix(strings.ToLower(m.Content), "!steam report") {
		urlIndex = 2
	}

	if len(parts) < urlIndex+1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!report <steam_profile_url> [number_of_bots]`\nExample: `!report https://steamcommunity.com/profiles/123456 5`")
		return
	}

	target := parts[urlIndex]
	numBots := 1
	if len(parts) >= urlIndex+2 {
		fmt.Sscanf(parts[urlIndex+1], "%d", &numBots)
		if numBots < 1 {
			numBots = 1
		}
	}

	var steamID string
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		var err error
		steamID, err = extractSteamID(target)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid Steam profile URL: %v", err))
			return
		}
	} else {
		// Assume direct SteamID input
		steamID = target
	}

	// Check if we have enough bot accounts
	availableBots := manager.ListCredentials()
	if len(availableBots) < numBots {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not enough bot accounts. Requested: %d, Available: %d\nUse `!steam bot-add` to add more accounts", numBots, len(availableBots)))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Valid Steam profile detected\nTarget Steam ID: `%s`\nDeploying **%d** bot(s)...", steamID, numBots))

	reportURL := fmt.Sprintf("https://help.steampowered.com/en/wizard/HelpWithGameIssue/?appid=730&issueid=1010&playerid=%s", steamID)

	var sessionIDs []string
	var botNames []string
	var successList []string

	// Create sessions for each bot
	for i := 0; i < numBots && i < len(availableBots); i++ {
		botName := availableBots[i]

		go func(idx int, account string) {
			session, err := manager.AddSession(account)
			if err != nil {
				log.Printf("Error creating session for %s: %v", account, err)
				return
			}
			log.Printf("Session %s created for bot %s", session.ID, account)
		}(i, botName)

		sessionID := fmt.Sprintf("session-%d", i+1)
		port := 9222 + i
		sessionIDs = append(sessionIDs, sessionID)
		botNames = append(botNames, botName)
		successList = append(successList, fmt.Sprintf("• **%s** → Session `%s` (port %d)", botName, sessionID, port))
	}

	time.Sleep(2 * time.Second) // Give sessions time to start

	summary := fmt.Sprintf("\n**Report Summary:**\nDeploying: %d/%d bots\n", numBots, numBots)
	summary += "\n**Active Bots:**\n"
	for _, bot := range successList {
		summary += bot + "\n"
	}
	summary += fmt.Sprintf("\nReport URL: `%s`\n", reportURL)
	summary += "Sessions starting - bots will login and navigate to report page\n\n_Automation in progress_"

	s.ChannelMessageSend(m.ChannelID, summary)

	// Track the report
	reportsMutex.Lock()
	reportID := fmt.Sprintf("report-%d", reportCounter)
	reportCounter++

	report := &ReportProcess{
		ReportID:   reportID,
		TargetID:   steamID,
		ProfileURL: target,
		SessionIDs: sessionIDs,
		BotNames:   botNames,
		StartTime:  time.Now(),
		Status:     "active",
	}

	reports[reportID] = report
	reportsMutex.Unlock()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\nReport tracked as: `%s`", reportID))
	go cleanupMessages(s, m.ChannelID, 10)
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

func handleReports(s *discordgo.Session, m *discordgo.MessageCreate) {
	reportsMutex.RLock()
	defer reportsMutex.RUnlock()

	if len(reports) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No active reports")
		return
	}

	response := "**Active Report Processes:**\n\n"

	for _, report := range reports {
		elapsed := time.Since(report.StartTime).Round(time.Second)
		response += fmt.Sprintf("**%s**\n", report.ReportID)
		response += fmt.Sprintf("Target: `%s`\n", report.TargetID)
		response += fmt.Sprintf("URL: %s\n", report.ProfileURL)
		response += fmt.Sprintf("Bots: %d active\n", len(report.BotNames))
		response += fmt.Sprintf("Running: %s\n", elapsed)
		response += fmt.Sprintf("Status: %s\n", report.Status)
		response += "```\n"
		for i, botName := range report.BotNames {
			response += fmt.Sprintf("%d. %s -> %s\n", i+1, botName, report.SessionIDs[i])
		}
		response += "```\n\n"
	}

	s.ChannelMessageSend(m.ChannelID, response)
	go cleanupMessages(s, m.ChannelID, 5)
}

func handleSteamHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	help := `**Steam Report Bot Commands**

**Bot Management:**
` + "`!steam bot-add <email> <username> <password> [shared_secret]`" + ` - Add bot account
` + "`!steam bot-list`" + ` - List all bot accounts
` + "`!steam bot-remove <username>`" + ` - Remove bot account

**Actions:**
` + "`!steam report <steam_profile_url> [num_bots]`" + ` - Report a player
` + "`!steam reports`" + ` - View active reports
` + "`!steam help`" + ` - Show this help message

_All commands work in DMs only for privacy_`

	s.ChannelMessageSend(m.ChannelID, help)
	go cleanupMessages(s, m.ChannelID, 3)
}

// GetManager returns the global session manager
func GetManager() *SessionManager {
	return manager
}
