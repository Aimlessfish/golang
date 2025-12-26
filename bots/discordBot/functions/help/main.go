package help

import "github.com/bwmarrin/discordgo"

func DisplayHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title:       "Bot Help",
		Description: "All Bot Commands: ",
		Color:       0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "\u200B", Value: "_**General Commands:**_", Inline: false},
			{Name: "/help", Value: "Displays this message!"},
			{Name: "/clear", Value: "Clears _100_ bot sent messages!"},
			{Name: "/football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
			{Name: "/proxy", Value: "Sends 1, tested; working, HTTP proxy."},
			// --- Generators ---
			{Name: "\u200B", Value: "**Generators:**", Inline: false},
			{Name: "/number <length>", Value: "Generates a random number with the specified number of digits (1-18). Example: !number 5"},
			{Name: "/username <input>", Value: "Generates a realistic username based on your input. Example: !username JohnDoe"},
			{Name: "/string <length>", Value: "Generates a random string with the specified length. Example: !string 8"},
			// --- Temp Mail ---
			{Name: "\u200B", Value: "**Temp Mail Commands:**", Inline: false},
			{Name: "/yopmail", Value: "Get a new YOPmail email address."},
			{Name: "/mail", Value: "Get a new GuerrillaMail email address."},
			{Name: "/inbox <INBOX_TOKEN>", Value: "Show the GuerrillaMail inbox for your token. \nExample: /inbox <INBOX_TOKEN>"},

			// --- Steam Bot Commands ---
			{Name: "\u200B", Value: "_**Steam Bot Commands:**_", Inline: false},
			{Name: "/help steam", Value: "Displays Steam bot commands."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}

func DisplayHelpSteam(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title:       "Steam Bot Help",
		Description: "Steam Bot Commands: ",
		Color:       0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/bot-add <email> <username> <password> [shared_secret]", Value: "Add a Steam bot account for reporting."},
			{Name: "/bot-list", Value: "List all added Steam bot accounts."},
			{Name: "/bot-remove <username>", Value: "Remove a Steam bot account by username."},
			{Name: "/bot-del <username>", Value: "Remove a Steam bot account by username."},
			{Name: "/report <url/steam64> [amount]", Value: "Report a player using the specified number of bots (default 1)."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
