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
			{Name: "\u200B", Value: "_**Betting Commands:**_", Inline: false},
			//Util help
			{Name: "/help bets", Value: "Displays all betting commands."},
			{Name: "\u200B", Value: "_**Utility Commands:**_", Inline: false},
			{Name: "/help util", Value: "Displays utility commands."},
			//Generators Help
			{Name: "\u200B", Value: "**Generator Commands:**", Inline: false},
			{Name: "/help generators", Value: "Displays Generator commands"}, // Temp Mail Help
			{Name: "\u200B", Value: "_**Temp Mail Commands:**_", Inline: false},
			{Name: "/help mail", Value: "Displays Temp Mail commands."},
			// Steam Bot help
			{Name: "\u200B", Value: "_**Steam Commands:**_", Inline: false},
			{Name: "/help steam", Value: "Displays Steam bot commands."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}

func DisplayHelpSteam(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Steam Bot Commands:",
		Color: 0x00ffcc,
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

func DisplayHelpMail(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Temp Mail Commands:",
		Color: 0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/yopmail", Value: "Get a new YOPmail email address."},
			{Name: "/mail", Value: "Get a new GuerrillaMail email address."},
			{Name: "/inbox <token>", Value: "Show the GuerrillaMail inbox for your token. \nExample: /inbox <INBOX_TOKEN>"},
			{Name: "/view <mail_id> <token>", Value: "View a specific email from your GuerrillaMail inbox using the mail ID and your token. \nExample: /view <MAIL_ID> <INBOX_TOKEN>"},
			{Name: "/del <mail_id> <token>", Value: "Delete a specific email from your GuerrillaMail inbox using the mail ID and your token. \nExample: /del <MAIL_ID> <INBOX_TOKEN>"},
			{Name: "/address <token>", Value: "Get the GuerrillaMail email address associated with your token. \nExample: /address <INBOX_TOKEN>"},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}

func DisplayUtilityHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Utility Commands:",
		Color: 0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/clear", Value: "Clears _100_ bot sent messages!"},
			{Name: "/proxy", Value: "Sends 1, tested; working, HTTP proxy."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil

}

func DisplayBettingHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Betting Commands:",
		Color: 0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}

func DisplayHelpGenerators(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Generator Commands:",
		Color: 0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/number <length>", Value: "Generates a random number with the specified number of digits (1-18). Example: !number 5"},
			{Name: "/username <input>", Value: "Generates a realistic username based on your input. Example: !username JohnDoe"},
			{Name: "/string <length>", Value: "Generates a random string with the specified length. Example: !string 8"},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
