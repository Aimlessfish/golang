package bot

import (
	"encoding/json"
	"html"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	betting "discordBot/functions/betting"
	clear "discordBot/functions/clearbotmsg"
	"discordBot/functions/generators"
	"discordBot/functions/help"
	getproxy "discordBot/functions/proxy"
	"discordBot/functions/servercheck"
	steammarket "discordBot/functions/steamMarket"
	"discordBot/functions/tempmail"
	util "discordBot/util"

	"github.com/bwmarrin/discordgo"
)

const (
	STEAM_URL = "https://steamcommunity.com/market/priceoverview/?appid=730&currency=3&market_hash_name="
)
const allowedChannelID = "1458239504698704037" // Replace with your specific channel ID

func ConnectAPI(logger *slog.Logger) error {
	logger = logger.With("Bot", "ConnectAPI")

	// Load environment variables
	if err := util.LoadEnv(); err != nil {
		logger.Warn("No .env file found, using defaults")
	}

	api_key := util.GetToken()

	discord, err := discordgo.New("Bot " + api_key)
	if err != nil {
		logger.Error("API connect failed!")
		return err
	}
	err = discord.Open()
	if err != nil {
		panic(err)
	}
	discord.AddHandler(messageHandler)
	logger.Info("Bot is running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	return nil
}

func messageHandler(server *discordgo.Session, message *discordgo.MessageCreate) {
	userID := message.Author.ID
	channelID := message.ChannelID

	if message.Author.ID == server.State.User.ID {
		return
	}

	// Only respond in DMs or the specific allowed channel
	if message.GuildID != "" && channelID != allowedChannelID {
		return
	} else if message.GuildID == "" || message.ChannelID == allowedChannelID {
		if strings.HasPrefix(message.Content, "/") {
			server.ChannelMessageSendReply(message.ChannelID, "loading..", &discordgo.MessageReference{
				MessageID: message.ID,
				ChannelID: message.ChannelID,
				GuildID:   message.GuildID,
			})

			if strings.HasPrefix(message.Content, "/help") || strings.HasPrefix(message.Content, "help") || strings.HasPrefix(message.Content, "commands") {
				//chop the string get the argument and call the steam helo function
				switch {
				case strings.HasPrefix(message.Content, "/help steam"), strings.HasPrefix(message.Content, "help steam"):
					help.DisplayHelpSteam(channelID, server, message)
					return
				case strings.HasPrefix(message.Content, "/help mail"), strings.HasPrefix(message.Content, "help mail"):
					help.DisplayHelpMail(channelID, server, message)
					return
				case strings.HasPrefix(message.Content, "/help util"), strings.HasPrefix(message.Content, "help utility"):
					help.DisplayUtilityHelp(channelID, server, message)
					return
				case strings.HasPrefix(message.Content, "/help bets"), strings.HasPrefix(message.Content, "help betting"):
					help.DisplayBettingHelp(channelID, server, message)
					return
				case strings.HasPrefix(message.Content, "/help generators"), strings.HasPrefix(message.Content, "help generators"):
					help.DisplayHelpGenerators(channelID, server, message)
					return
				default:
					help.DisplayHelp(channelID, server, message)
				}
			}
			if message.Content == "/proxy" {
				proxies := getproxy.ProxyHandler(1)
				for _, proxy := range proxies {
					server.ChannelMessageSend(message.ChannelID, proxy)
				}
			}
			if message.Content == "/clear" {
				v := clear.ClearBotMessages(userID, channelID, server, message)
				if !v {
					server.ChannelMessageSend(message.ChannelID, "failed to clear messages!")
				}
			}
			if message.Content == "/football" {
				err := betting.MatchOdds(server, message)
				if err != nil {
					server.ChannelMessageSend(channelID, "Failed to retrieve upcoming matches! ")
				}
			}
			/***Steam Report Functions***/
			// Main Report Message Handler
			if strings.HasPrefix(message.Content, "/report") || strings.HasPrefix(message.Content, "report") {
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
					util.ExecBinary("./bin/csreport", uid, amount)
				} else {
					server.ChannelMessageSend(message.ChannelID, "Report started for: \n (uid: "+uid+")")
					util.ExecBinary("./bin/csreport", uid, "1")
				}
			}
			// Bot Addition Handler
			if strings.HasPrefix(message.Content, "/bot-add") || strings.HasPrefix(message.Content, "bot-add") {
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
				output, err := util.ExecBinary("./bin/csreport", command, args...)
				if err != nil {
					server.ChannelMessageSend(message.ChannelID, "Failed to add bot account!")
				} else {
					server.ChannelMessageSend(message.ChannelID, "\n"+output)
				}
			}
			// Bot Removal HAndler
			if strings.HasPrefix(message.Content, "/bot-remove") || strings.HasPrefix(message.Content, "bot-remove") || strings.HasPrefix(message.Content, "/bot-del") || strings.HasPrefix(message.Content, "bot-del") {
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
				output, err := util.ExecBinary("./bin/csreport", command, args...)
				if err != nil {
					server.ChannelMessageSend(message.ChannelID, "Failed to remove bot account!")
				} else {
					server.ChannelMessageSend(message.ChannelID, "\n"+output)
				}
			}
			// List Handler
			if strings.HasPrefix(message.Content, "/bot-list") || strings.HasPrefix(message.Content, "bot-list") {
				parts := util.SplitArgs(message.Content)
				if len(parts) < 1 {
					server.ChannelMessageSend(message.ChannelID, "Usage: /bot-list")
					return
				}
				command := "bot-list"
				args := []string{}
				output, err := util.ExecBinary("./bin/reporter/csreport", command, args...)
				if err != nil {
					server.ChannelMessageSend(message.ChannelID, "Failed to list bot accounts!")
				} else {
					server.ChannelMessageSend(message.ChannelID, "\n"+"```"+output+"```")
				}
			}
			/***Generators***/
			if strings.HasPrefix(message.Content, "/number ") {
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

			if strings.HasPrefix(message.Content, "/username ") {
				parts := strings.Split(message.Content, " ")
				if len(parts) != 2 {
					server.ChannelMessageSend(channelID, "Please provide a valid input for the username. Example: /username JohnDoe")
					return
				}
				input := parts[1]
				username := generators.GenerateUsername(input)
				server.ChannelMessageSend(channelID, "```Generated Username: "+username+"```")
			}
			if strings.HasPrefix(message.Content, "/string") {
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
			// email handlers
			if strings.HasPrefix(message.Content, "/yopmail") {
				email, domains, err := tempmail.GetRandomYopmail()
				parts := strings.Split(email, "'")
				prefix := parts[1]
				url := "https://yopmail.com/en/inbox?login=" + prefix
				if err != nil {
					server.ChannelMessageSend(channelID, "Failed to generate random email.")
				} else {
					server.ChannelMessageSend(channelID, "```Email: "+prefix+"\nInbox: "+url+"\n"+"Alternate Domains:\n"+domains+"```")
				}
			}
			if strings.HasPrefix(message.Content, "/mail") {
				email, sidToken, err := tempmail.GetRandomGuerrillaEmail()
				if err != nil {
					server.ChannelMessageSend(channelID, "Failed to generate random guerrilla email.")
				} else {
					server.ChannelMessageSend(channelID, "```Email: "+email+"\nInbox Token: "+sidToken+"\n *Keep your token safe to access your inbox!*```")
				}
			}
			if strings.HasPrefix(message.Content, "/inbox") {
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
				output, err := tempmail.GetGuerrillaInboxRaw(sidToken) // This should return a Go struct/slice
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
			// view email content handler
			if strings.HasPrefix(message.Content, "/view") {
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
			// delete email handler
			if strings.HasPrefix(message.Content, "/del") {
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
			if strings.HasPrefix(message.Content, "/address") {
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
			} // End of email handlers
			if strings.HasPrefix(message.Content, "/servers") {
				output, err := servercheck.CheckServers()
				if err != nil {
					server.ChannelMessageSend(channelID, "Failed to check server status: "+err.Error())
					return
				}
				server.ChannelMessageSend(channelID, output)

			}
		} else {
			server.ChannelMessageSend(message.ChannelID, "Send me a DM to use commands ")
		}
	}
	/*** Steam Market Command Handlers ***/

	// get market item info
	if strings.HasPrefix(message.Content, "/price") || strings.HasPrefix(message.Content, "price") {
		parts := util.SplitArgs(message.Content)
		if len(parts) < 2 {
			server.ChannelMessageSend(message.ChannelID, "Usage: /price AK-47 | Redline (Field-Tested)")
			return
		}
		itemName := strings.Join(parts[1:], " ")
		if itemName == "" {
			server.ChannelMessageSend(message.ChannelID, "you did not provide a valid item name")
			return
		}

		if strings.HasPrefix(message.Content, "!number ") {
			parts := strings.Split(message.Content, " ")
			if len(parts) != 2 {
				server.ChannelMessageSend(channelID, "Please provide a valid length for the random number. Example: !number 5")
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
			server.ChannelMessageSend(channelID, "Generated Random Number: "+randomNumber)
		}

		if strings.HasPrefix(message.Content, "!username ") {
			parts := strings.Split(message.Content, " ")
			if len(parts) != 2 {
				server.ChannelMessageSend(channelID, "Please provide a valid input for the username. Example: !username JohnDoe")
				return
			}
			input := parts[1]
			username := generators.GenerateUsername(input)
			server.ChannelMessageSend(channelID, "Generated Username: "+username)
		}
		if strings.HasPrefix(message.Content, "!string") {
			parts := strings.Split(message.Content, " ")
			if len(parts) != 2 {
				server.ChannelMessageSend(channelID, "Please provide a valid length for the random string. Example: !string 10")
				return
			}
			length, err := strconv.Atoi(parts[1])
			if err != nil || length <= 0 {
				server.ChannelMessageSend(channelID, "Please provide a valid positive integer for the length.")
				return
			}
			randomString := generators.GenerateRandomString(length)
			server.ChannelMessageSend(channelID, "Generated Random String: "+randomString)
		}
		// Call SteamMarketPriceOverviewURL to get price data
		var priceOutput string
		var err error

		// Try to call the steammarket package function
		priceOutput, err = steammarket.SteamMarketPriceOverviewURL(itemName)
		if err != nil {
			server.ChannelMessageSend(message.ChannelID, "Error fetching price: "+err.Error())
			return
		}
		server.ChannelMessageSend(message.ChannelID, "\n"+priceOutput)
		return
	}

}
