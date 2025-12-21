package help

import "github.com/bwmarrin/discordgo"

func DisplayHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title:       "Bot Help",
		Description: "All Bot Commands: ",
		Color:       0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "!help", Value: "Displays this message!"},
			{Name: "!clear", Value: "Clears _100_ bot sent messages!"},
			{Name: "!football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
			{Name: "!proxy", Value: "Sends 1, tested; working, HTTP proxy."},
			// --- Steam Bot Commands ---
			{Name: "\u200B", Value: "**Steam Bot Commands:**", Inline: false},
			{Name: "!steam bot-add <email> <username> <password> [shared_secret]", Value: "Add a Steam bot account for reporting."},
			{Name: "!steam bot-list", Value: "List all added Steam bot accounts."},
			{Name: "!steam bot-remove <username>", Value: "Remove a Steam bot account by username."},
			{Name: "!report <steam_profile_url> [num_bots]", Value: "Report a player using the specified number of bots (default 1)."},
			{Name: "!steam reports", Value: "View active report jobs."},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
