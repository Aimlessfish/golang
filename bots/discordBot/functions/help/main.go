package help

import "github.com/bwmarrin/discordgo"

func DisplayHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title:       "Bot Help",
		Description: "All Bot Commands: ",
		Color:       0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/help", Value: "Displays this message!"},
			{Name: "/clear", Value: "Clears _100_ bot sent messages!"},
			{Name: "/football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
			{Name: "/proxy", Value: "Sends 1, tested; working, HTTP proxy."},
			// --- Steam Bot Commands ---
			{Name: "\u200B", Value: "**Steam Bot Commands:**", Inline: false},
			{Name: "/bot-add <email> <username> <password> [shared_secret]", Value: "Add a Steam bot account for reporting."},
			{Name: "/bot-list", Value: "List all added Steam bot accounts."},
			{Name: "/bot-remove <username>", Value: "Remove a Steam bot account by username."},
			{Name: "/bot-del <username>", Value: "Remove a Steam bot account by username."},
			{Name: "/report <url/steam64> [amount]", Value: "Report a player using the specified number of bots (default 1)."},
			{Name: "/number <length>", Value: "Generates a random number with the specified number of digits (1-18). Example: !number 5"},
			{Name: "/username <input>", Value: "Generates a realistic username based on your input. Example: !username JohnDoe"},
			{Name: "/string <length>", Value: "Generates a random string with the specified length. Example: !string 8"},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
