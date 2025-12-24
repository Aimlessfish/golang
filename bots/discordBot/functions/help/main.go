package help

import "github.com/bwmarrin/discordgo"

func DisplayHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title: "Bot Help",
		Color: 0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			// --- General Commands ---
			{Name: "\u200B", Value: "**General Commands:**", Inline: false},
			{Name: "!help", Value: "Displays this message!"},
			{Name: "!clear", Value: "Clears _100_ bot sent messages!"},
			// --- Useless Commands ---
			{Name: "\u200B", Value: "**Useless Commands:**", Inline: false},
			{Name: "!football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
			{Name: "!proxy", Value: "Sends 1, tested; working, HTTP proxy."},
			// --- Steam Report Commands ---
			{Name: "\u200B", Value: "**Steam Report Commands:**", Inline: false},
			{Name: "!bot-add <email> <username> <password> [shared_secret]\n *2FA MUST BE DISABLED*", Value: "Add a Steam bot account for reporting."},
			{Name: "!bot-list", Value: "List all added Steam bot accounts."},
			{Name: "!bot-remove <username>", Value: "Remove a Steam bot account by username."},
			{Name: "!bot-del <username>", Value: "Remove a Steam bot account by username."},
			{Name: "!report <url/steam64> [amount]", Value: "Report a player using the specified number of bots (default 1)."},
			// --- Steam Market Commands ---
			{Name: "\u200B", Value: "**Steam Market Commands:**", Inline: false},
			{Name: "!price <item_name>", Value: "Fetches the current market price for the specified item."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
