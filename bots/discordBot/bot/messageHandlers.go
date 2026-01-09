package bot

import (
	"encoding/json"
	"html"
	"regexp"
	"strconv"
	"strings"

	betting "discordBot/functions/betting"
	clear "discordBot/functions/clearbotmsg"
	"discordBot/functions/generators"
	"discordBot/functions/help"
	getproxy "discordBot/functions/proxy"
	"discordBot/functions/tempmail"
	"discordBot/util"

	"github.com/bwmarrin/discordgo"
)

const (
	CS_REPORTER_BINARY     = "../bin/cs_reporter"
	SERVICE_CHECKER_BINARY = "../bin/service_checker"
)

// Handler function signatures
func HandleDM(server *discordgo.Session, message *discordgo.MessageCreate) {
	userID := message.Author.ID
	dmChannel, err := server.UserChannelCreate(userID)
	if err != nil {
		server.ChannelMessageSend(message.ChannelID, "Failed to open DM: "+err.Error())
		return
	}
	server.ChannelMessageSend(dmChannel.ID, "Hello! This is your DM with the bot. You can interact with me here. \n Run /help to see available commands.")
	server.ChannelMessageSend(message.ChannelID, "I've sent you a DM!")
}
func HandleHelp(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	// chop the string get the argument and call the steam help function
	content := message.Content
	switch {
	case strings.HasPrefix(content, "/help steam"), strings.HasPrefix(content, "help steam"):
		help.DisplayHelpSteam(channelID, server, message)
	case strings.HasPrefix(content, "/help mail"), strings.HasPrefix(content, "help mail"):
		help.DisplayHelpMail(channelID, server, message)
	case strings.HasPrefix(content, "/help util"), strings.HasPrefix(content, "help utility"):
		help.DisplayUtilityHelp(channelID, server, message)
	case strings.HasPrefix(content, "/help bets"), strings.HasPrefix(content, "help betting"):
		help.DisplayBettingHelp(channelID, server, message)
	case strings.HasPrefix(content, "/help generators"), strings.HasPrefix(content, "help generators"):
		help.DisplayHelpGenerators(channelID, server, message)
	default:
		help.DisplayHelp(channelID, server, message)
	}
}
func HandleProxy(server *discordgo.Session, message *discordgo.MessageCreate) {
	parts := util.SplitArgs(message.Content)
	proxyType := "http" // default
	if len(parts) > 1 {
		switch parts[1] {
		case "http", "https", "socks5":
			proxyType = parts[1]
		default:
			server.ChannelMessageSend(message.ChannelID, "Invalid proxy type. Use: http, https, or socks5")
			return
		}
	}
	proxies := getproxy.ProxyHandler(proxyType)
	for _, proxy := range proxies {
		server.ChannelMessageSend(message.ChannelID, proxy)
	}
}
func HandleClear(server *discordgo.Session, message *discordgo.MessageCreate) {
	userID := message.Author.ID
	channelID := message.ChannelID
	v := clear.ClearBotMessages(userID, channelID, server, message)
	if !v {
		server.ChannelMessageSend(message.ChannelID, "failed to clear messages!")
	}
}
func HandleFootball(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	err := betting.MatchOdds(server, message)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to retrieve upcoming matches! ")
	}
}
func HandleReport(server *discordgo.Session, message *discordgo.MessageCreate) {
	parts := util.SplitArgs(message.Content)
	if len(parts) < 2 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /report <uid> <amount>")
		return
	}
	uid := parts[1]
	if uid == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a steam64 ID or a valid profile URL")
		return
	}
	if len(parts) >= 3 {
		amount := parts[2]
		server.ChannelMessageSend(message.ChannelID, amount+" Reports started for: \n (uid: "+uid+")")
		util.ExecBinary(uid, amount)
	} else {
		server.ChannelMessageSend(message.ChannelID, "Report started for: \n (uid: "+uid+")")
		util.ExecBinary(uid, "1")
	}
}
func HandleBotAdd(server *discordgo.Session, message *discordgo.MessageCreate) {
	parts := util.SplitArgs(message.Content)
	if len(parts) < 3 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /bot-add <username> <password>")
		return
	}
	username := parts[1]
	password := parts[2]
	if username == "" || password == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid username or password")
		return
	}
	command := "add"
	args := []string{username, password}
	output, err := util.ExecBinary(CS_REPORTER_BINARY, command, args...)
	if err != nil {
		server.ChannelMessageSend(message.ChannelID, "Failed to add bot account!")
	} else {
		server.ChannelMessageSend(message.ChannelID, "\n"+output)
	}
}
func HandleBotRemove(server *discordgo.Session, message *discordgo.MessageCreate) {
	parts := util.SplitArgs(message.Content)
	if len(parts) < 2 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /bot-remove <username>")
		return
	}
	username := parts[1]
	if username == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid username")
		return
	}
	command := "bot-remove"
	args := []string{username}
	output, err := util.ExecBinary(CS_REPORTER_BINARY, command, args...)
	if err != nil {
		server.ChannelMessageSend(message.ChannelID, "Failed to remove bot account!")
	} else {
		server.ChannelMessageSend(message.ChannelID, "\n"+output)
	}
}
func HandleBotList(server *discordgo.Session, message *discordgo.MessageCreate) {
	command := "bot-list"
	args := []string{}
	output, err := util.ExecBinary(CS_REPORTER_BINARY, command, args...)
	if err != nil {
		server.ChannelMessageSend(message.ChannelID, "Failed to list bot accounts!")
	} else {
		server.ChannelMessageSend(message.ChannelID, "\n"+"```"+output+"```")
	}
}
func HandleNumber(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := strings.Split(message.Content, " ")
	if len(parts) != 2 {
		server.ChannelMessageSend(channelID, "Please provide a valid length for the random number. Example: /number 5")
		return
	}
	input := parts[1]
	length, err := strconv.Atoi(input)
	if err != nil || length <= 0 {
		server.ChannelMessageSend(channelID, "Please provide a valid positive integer for the length.")
		return
	}
	if length > 18 {
		server.ChannelMessageSend(channelID, "Please provide a number length between 1 and 18.")
		return
	}
	randomNumber := generators.GenerateRandomNumber(input)
	server.ChannelMessageSend(channelID, "```Generated Random Number: "+randomNumber+"```")
}
func HandleUsername(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := strings.Split(message.Content, " ")
	if len(parts) != 2 {
		server.ChannelMessageSend(channelID, "Please provide a valid input for the username. Example: /username JohnDoe")
		return
	}
	input := parts[1]
	username := generators.GenerateUsername(input)
	server.ChannelMessageSend(channelID, "```Generated Username: "+username+"```")
}
func HandleString(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := strings.Split(message.Content, " ")
	if len(parts) != 2 {
		server.ChannelMessageSend(channelID, "Please provide a valid length for the random string. Example: /string 10")
		return
	}
	length, err := strconv.Atoi(parts[1])
	if err != nil || length <= 0 {
		server.ChannelMessageSend(channelID, "Please provide a valid positive integer for the length.")
		return
	}
	randomString := generators.GenerateRandomString(length)
	server.ChannelMessageSend(channelID, "```Generated Random String: "+randomString+"```")
}
func HandleYopmail(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	email, domains, err := tempmail.GetRandomYopmail()
	parts := strings.Split(email, "'")
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[1]
	}
	url := "https://yopmail.com/en/inbox?login=" + prefix
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to generate random email.")
	} else {
		server.ChannelMessageSend(channelID, "```Email: "+prefix+"\nInbox: "+url+"\n"+"Alternate Domains:\n"+domains+"```")
	}
}
func HandleMail(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	email, sidToken, err := tempmail.GetRandomGuerrillaEmail()
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to generate random guerrilla email.")
	} else {
		server.ChannelMessageSend(channelID, "```Email: "+email+"\nInbox Token: "+sidToken+"\n *Keep your token safe to access your inbox!*```")
	}
}
func HandleInbox(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := util.SplitArgs(message.Content)
	if len(parts) < 2 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /inbox <sid_token>")
		return
	}
	sidToken := parts[1]
	if sidToken == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid sid_token")
		return
	}
	output, err := tempmail.GetGuerrillaInboxRaw(sidToken)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to get inbox: "+err.Error())
		return
	}
	var resp tempmail.GuerrillaInboxResponse
	err = json.Unmarshal([]byte(output), &resp)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to parse inbox JSON: "+err.Error())
		return
	}
	if len(resp.List) == 0 {
		server.ChannelMessageSend(channelID, "No emails found in inbox.")
		return
	}
	var msg strings.Builder
	msg.WriteString("***inbox:***\n")
	for _, mail := range resp.List {
		msg.WriteString("```MailID: " + mail.MailID +
			" | From: " + mail.MailFrom +
			" | Subject: " + mail.MailSubject + "```\n")
	}
	server.ChannelMessageSend(channelID, msg.String())
}
func HandleView(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := util.SplitArgs(message.Content)
	if len(parts) < 3 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /view <sid_token> <mail_id>")
		return
	}
	sidToken := parts[1]
	mailID := parts[2]
	if sidToken == "" || mailID == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid sid_token or mail_id")
		return
	}
	output, err := tempmail.GetGuerrillaMailContent(sidToken, mailID)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to get email content: "+err.Error())
		return
	}

	// Check for boolean or error response
	var boolCheck bool
	if err := json.Unmarshal([]byte(output), &boolCheck); err == nil {
		server.ChannelMessageSend(channelID, "No email found or invalid response from API.")
		return
	}

	// Try to unmarshal as email content
	var emailContent struct {
		MailID      string `json:"mail_id"`
		MailFrom    string `json:"mail_from"`
		MailSubject string `json:"mail_subject"`
		MailExcerpt string `json:"mail_excerpt"`
		MailBody    string `json:"mail_body"`
	}
	err = json.Unmarshal([]byte(output), &emailContent)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to parse email content JSON: "+err.Error())
		return
	}
	body := emailContent.MailBody
	// Replace <br> and <br/> with newlines
	body = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(body, "\n")
	// Remove all other HTML tags
	body = regexp.MustCompile(`(?s)<.*?>`).ReplaceAllString(body, "")
	// Decode HTML entities
	body = html.UnescapeString(body)

	// Remove empty lines from the body
	lines := strings.Split(body, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	cleanBody := strings.Join(nonEmptyLines, "\n")

	msg := "```***Email Content:***\n"
	msg += "From: " + emailContent.MailFrom + "\n"
	msg += "Subject: " + emailContent.MailSubject + "\n"
	msg += "Body:\n" + cleanBody + "```"
	server.ChannelMessageSend(channelID, msg)
}
func HandleDel(server *discordgo.Session, message *discordgo.MessageCreate) {
	channelID := message.ChannelID
	parts := util.SplitArgs(message.Content)
	if len(parts) < 3 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /del <sid_token> <mail_id>")
		return
	}
	mailID := parts[1]
	sidToken := parts[2]
	if sidToken == "" || mailID == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid sid_token or mail_id")
		return
	}
	output, err := tempmail.DeleteGuerrillaMail(mailID, sidToken)
	if err != nil {
		server.ChannelMessageSend(channelID, "Failed to delete email: "+err.Error())
		return
	}
	// Respond with 'deleted' if the API response is a valid JSON array or object (success)
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		server.ChannelMessageSend(channelID, "```deleted```")
		return
	}
	var resp interface{}
	err = json.Unmarshal([]byte(trimmed), &resp)
	if err == nil {
		server.ChannelMessageSend(channelID, "```deleted```")
	} else {
		server.ChannelMessageSend(channelID, "Delete API returned: "+trimmed)
	}
}
func HandleAddress(server *discordgo.Session, message *discordgo.MessageCreate) {
	parts := util.SplitArgs(message.Content)
	if len(parts) < 2 {
		server.ChannelMessageSend(message.ChannelID, "Usage: /address <email_address>")
		return
	}
	emailAddress := parts[1]
	if emailAddress == "" {
		server.ChannelMessageSend(message.ChannelID, "you did not provide a valid email address")
		return
	}
	email_addr, err := tempmail.GetGuerrillaEmailAddress(emailAddress, "en")
	if err != nil {
		server.ChannelMessageSend(message.ChannelID, "Failed to get email address!")
	} else {
		server.ChannelMessageSend(message.ChannelID, "```\nEmail: "+email_addr+"```")
	}
}
